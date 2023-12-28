package biz

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicCommon/foundation/net"
	fu "github.com/muidea/magicCommon/foundation/util"

	"github.com/muidea/magieAgent/internal/config"
	"github.com/muidea/magieAgent/pkg/common"
)

func (s *Base) SwitchMariadb(serviceName string, cfgPtr *config.CfgItem) (err *cd.Result) {
	serviceInfo, serviceErr := s.queryLocalServiceInfo(serviceName, common.Mariadb)
	if serviceErr != nil {
		log.Errorf("SwitchMariadb failed, s.queryLocalServiceInfo service:%s error:%s", serviceName, serviceErr.Error())
		err = serviceErr
		return
	}

	if cfgPtr.LocalHost == cfgPtr.Status.Master {
		ev := event.NewEvent(common.SwitchToMasterStatus, s.ID(), common.MariadbModule, nil, serviceInfo)
		result := s.SendEvent(ev)
		err = result.Error()
		if err != nil {
			log.Errorf("SwitchMariadb failed, service:%s switch to master error:%s", serviceName, err.Error())
			return
		}
		err = s.restoreServiceEndpoint(serviceInfo)
		return
	}

	// 如果是由master切换至slave，这里还需要考虑是否需要全量的做一次数据恢复
	// 做全量数据恢复，需要先从对端备份数据，将其还原至本机后，再切换服务状态
	masterStatus, statusErr := s.verifyRemoteMasterMariadb(serviceInfo, cfgPtr.RemoteHost)
	if statusErr != nil {
		log.Errorf("SwitchMariadb to slave failed, s.verifyRemoteMasterMariadb service:%s error:%s", serviceName, statusErr.Error())
		err = statusErr
		return
	}

	masterInfo := &common.MasterInfo{
		Host:    masterStatus.Host,
		Port:    masterStatus.Port,
		LogFile: masterStatus.LogFile,
		LogPos:  masterStatus.LogPos,
	}

	ev := event.NewEvent(common.SwitchToSlaveStatus, s.ID(), common.MariadbModule, nil, serviceInfo)
	ev.SetData("master", masterInfo)
	result := s.SendEvent(ev)
	err = result.Error()
	if err != nil {
		log.Errorf("SwitchMariadb failed, service:%s switch to slave error:%s", serviceName, err.Error())
		return
	}
	err = s.changeServiceEndpoint(serviceInfo)
	return
}

func (s *Base) BackupMariadb(serviceName string) (ret string, err *cd.Result) {
	serviceInfo, serviceErr := s.queryLocalServiceInfo(serviceName, common.Mariadb)
	if serviceErr != nil {
		log.Errorf("BackupMariadb failed, s.queryLocalServiceInfo service:%s error:%s", serviceName, serviceErr.Error())
		err = serviceErr
		return
	}

	backupEvent := event.NewEvent(common.BackupDatabase, s.ID(), common.MariadbModule, nil, serviceInfo)
	result := s.SendEvent(backupEvent)
	resultVal, resultErr := result.Get()
	if resultErr != nil {
		log.Errorf("BackupMariadb failed, backup mariadb:%s error:%s", serviceInfo, resultErr.Error())
		err = resultErr
		return
	}

	ret = resultVal.(string)
	return
}

func (s *Base) RestoreMariadb(serviceName string) (err *cd.Result) {
	serviceInfo, serviceErr := s.queryLocalServiceInfo(serviceName, common.Mariadb)
	if serviceErr != nil {
		log.Errorf("RestoreMariadb failed, s.queryLocalServiceInfo service:%s error:%s", serviceName, serviceErr.Error())
		err = serviceErr
		return
	}

	err = s.restoreMariadbInternal(serviceInfo, config.GetRemoteHost())
	if err != nil {
		log.Errorf("RestoreMariadb failed, s.restoreMariadbInternal service:%s error:%s", serviceInfo, err.Error())
		return
	}

	return
}

func (s *Base) QueryLocalMariadbStatus(serviceName string) (masterStatus *common.MasterStatus, slaveStatus *common.SlaveStatus, err *cd.Result) {
	serviceInfo, serviceErr := s.queryLocalServiceInfo(serviceName, common.Mariadb)
	if serviceErr != nil {
		log.Errorf("QueryLocalMariadbStatus failed, s.queryLocalServiceInfo service:%s error:%s", serviceName, serviceErr.Error())
		err = serviceErr
		return
	}

	masterEvent := event.NewEvent(common.CheckMasterStatus, s.ID(), common.MariadbModule, nil, serviceInfo)
	result := s.SendEvent(masterEvent)
	resultVal, resultErr := result.Get()
	if resultErr != nil {
		log.Errorf("QueryLocalMariadbStatus failed, check service:%s master status error:%s", serviceInfo, resultErr.Error())
		err = resultErr
		return
	}
	masterStatus = resultVal.(*common.MasterStatus)

	slaveEvent := event.NewEvent(common.CheckSlaveStatus, s.ID(), common.MariadbModule, nil, serviceInfo)
	result = s.SendEvent(slaveEvent)
	resultVal, resultErr = result.Get()
	if resultErr != nil {
		log.Errorf("QueryLocalMariadbStatus failed, check service:%s slave status error:%s", serviceInfo, resultErr.Error())
		err = resultErr
		return
	}

	slaveStatus = resultVal.(*common.SlaveStatus)
	return
}

func (s *Base) verifyRemoteMasterMariadb(serviceInfo *common.ServiceInfo, remoteHost string) (ret *common.MasterStatus, err *cd.Result) {
	log.Infof("verifyRemoteMasterMariadb %v, remoteHost:%v", serviceInfo, remoteHost)

	// 如果对端同时处于master和slave状态，则说明对端节点还在接收数据，这里不能直接进行切换
	masterStatus, slaveStatus, statusErr := s.queryRemoteMariadbStatus(serviceInfo, remoteHost)
	if statusErr != nil {
		log.Errorf("verifyRemoteMasterMariadb to slave failed, s.queryRemoteMariadbStatus service:%s error:%s", serviceInfo.Name, statusErr.Error())
		err = statusErr
		return
	}

	if masterStatus == nil || !masterStatus.Enable {
		log.Errorf("verifyRemoteMasterMariadb to slave failed, illegal remote master service:%s, remote:%s", serviceInfo, remoteHost)
		err = cd.NewError(cd.UnExpected, "missing master or master is disable")
		return
	}
	if slaveStatus != nil && slaveStatus.Enable {
		log.Errorf("verifyRemoteMasterMariadb to slave failed, illegal remote master service:%s, remote:%s", serviceInfo, remoteHost)
		err = cd.NewError(cd.UnExpected, "remote service remain in slave mode")
		return
	}

	// 这里判断是否需要全量的从对端备份一份数据过来
	// 1、如果本地数据是空的，则必需要全量同步一份数据过来
	// 2、如果当前节点掉线时间超过binlog数据时间了，则必需要全量同步一份数据过来
	needSyncData := true
	localStatus, localErr := s.loadMasterStatus(serviceInfo)
	if localErr == nil {
		elapse := time.Now().UTC().UnixMilli() - localStatus.TimeStamp
		if elapse < int64(3*24*time.Hour/time.Millisecond) {
			needSyncData = false
		}
	}

	if needSyncData {
		err = s.restoreMariadbInternal(serviceInfo, remoteHost)
		if err != nil {
			log.Errorf("verifyRemoteMasterMariadb to slave failed, s.restoreMariadbInternal service:%s, remote:%s, error:%s", serviceInfo, remoteHost, err.Error())
			return
		}

		logFile, logPos, logErr := s.extractBinlog(serviceInfo)
		if logErr != nil {
			err = logErr
			log.Errorf("verifyRemoteMasterMariadb to slave failed, s.extractBinlog service:%s, remote:%s, error:%s", serviceInfo, remoteHost, logErr.Error())
			return
		}

		masterStatus.LogFile = logFile
		masterStatus.LogPos = logPos
		statusErr = s.saveMasterStatus(serviceInfo, masterStatus)
		if statusErr != nil {
			log.Errorf("verifyRemoteMasterMariadb to slave failed, s.saveMasterStatus service:%s error:%s", serviceInfo.Name, statusErr.Error())
		}
	}

	ret = masterStatus
	return
}

func (s *Base) queryRemoteMariadbStatus(serviceInfo *common.ServiceInfo, remoteHost string) (masterStatus *common.MasterStatus, slaveStatus *common.SlaveStatus, err *cd.Result) {
	log.Infof("queryRemoteMariadbStatus %v, remoteHost:%v", serviceInfo, remoteHost)

	values := url.Values{}
	values.Set(common.APIKey, config.GetSyncSecret())
	result := &common.QueryMariadbStatusResult{}
	urlVal, _ := url.ParseRequestURI(fmt.Sprintf("http://%s:%s", remoteHost, config.GetNodePort()))
	urlVal.Path = strings.Join([]string{urlVal.Path, common.ApiVersion, common.QueryMariadbStatus}, "")
	urlVal.Path = strings.ReplaceAll(urlVal.Path, ":id", fmt.Sprintf("%s", serviceInfo.Name))

	client := &http.Client{}
	defer client.CloseIdleConnections()
	_, httpErr := net.HTTPGet(client, urlVal.String(), result, values)
	if httpErr != nil {
		err = cd.NewError(cd.UnExpected, httpErr.Error())
		return
	}

	masterStatus = result.MasterStatus
	slaveStatus = result.SlaveStatus
	return
}

func (s *Base) backupMariadbFromRemote(serviceInfo *common.ServiceInfo, remoteHost string) (ret string, err *cd.Result) {
	log.Infof("backupMariadbFromRemote %v, remoteHost:%v", serviceInfo, remoteHost)

	values := url.Values{}
	values.Set(common.APIKey, config.GetSyncSecret())

	result := &common.BackupMariadbDataResult{}
	urlVal, _ := url.ParseRequestURI(fmt.Sprintf("http://%s:%s", remoteHost, config.GetNodePort()))
	urlVal.Path = strings.Join([]string{urlVal.Path, common.ApiVersion, common.BackupMariadbData}, "")
	urlVal.Path = strings.ReplaceAll(urlVal.Path, ":id", fmt.Sprintf("%s", serviceInfo.Name))

	_, httpErr := net.HTTPGet(&http.Client{}, urlVal.String(), result, values)
	if httpErr != nil {
		err = cd.NewError(cd.UnExpected, httpErr.Error())
		return
	}

	ret = result.DataPath
	return
}

func (s *Base) syncRemoteMariadbToLocal(serviceInfo *common.ServiceInfo, remotePath string) (err *cd.Result) {
	syncInfo := &common.SyncInfo{
		Name:   serviceInfo.Name,
		Local:  serviceInfo.Volumes.BackPath.Value[len(config.GetTmpPath())+1:],
		Remote: remotePath[len(config.GetTmpPath())+1:],
	}

	syncEvent := event.NewEvent(common.SyncFilesToLocal, s.ID(), common.SyncthingModule, nil, syncInfo)
	result := s.SendEvent(syncEvent)
	err = result.Error()
	return
}

func (s *Base) restoreMariadbInternal(serviceInfo *common.ServiceInfo, remoteHost string) (err *cd.Result) {
	log.Infof("restoreMariadbInternal %v, remoteHost:%v", serviceInfo, remoteHost)

	dataPath, dataErr := s.backupMariadbFromRemote(serviceInfo, remoteHost)
	if dataErr != nil {
		err = dataErr
		log.Errorf("verifyRemoteMasterMariadb to slave failed, s.backupMariadbFromRemote service:%s, remote:%s, error:%s", serviceInfo, remoteHost, dataErr.Error())
		return
	}
	dataErr = s.syncRemoteMariadbToLocal(serviceInfo, dataPath)
	if dataErr != nil {
		err = dataErr
		log.Errorf("verifyRemoteMasterMariadb to slave failed, s.backupMariadbFromRemote service:%s, remote:%s, error:%s", serviceInfo, remoteHost, dataErr.Error())
		return
	}

	restoreEvent := event.NewEvent(common.RestoreDatabase, s.ID(), common.MariadbModule, nil, serviceInfo)
	result := s.SendEvent(restoreEvent)
	resultErr := result.Error()
	if resultErr != nil {
		log.Errorf("verifyRemoteMasterMariadb failed, restore mariadb:%s error:%s", serviceInfo, resultErr.Error())
		err = resultErr
		return
	}

	return
}

func (s *Base) extractBinlog(serviceInfo *common.ServiceInfo) (logFile string, logPos int64, err *cd.Result) {
	log.Infof("extractBinlog service:%v", serviceInfo)

	binlogFile := path.Join(serviceInfo.Volumes.BackPath.Value, common.MariadbBackupBinlogFile)
	filePtr, fileErr := os.OpenFile(binlogFile, os.O_RDONLY, os.ModePerm)
	if fileErr != nil {
		err = cd.NewError(cd.UnExpected, fileErr.Error())
		log.Warnf("extractBinlog open binlog file error:%s", fileErr.Error())
		return
	}
	defer filePtr.Close()

	scanner := bufio.NewScanner(filePtr)
	if scanner.Scan() {
		values := scanner.Text()
		items := strings.Split(values, "\t")
		if len(items) != 3 {
			err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal bin log file, itemSize:%v, %v", len(items), items))
			return
		}
		logFile = items[0]
		logPos, _ = strconv.ParseInt(items[1], 10, 64)
		return
	}

	err = cd.NewError(cd.UnExpected, "bin log file is empty")
	return
}

func (s *Base) loadSlaveStatus(serviceInfo *common.ServiceInfo) (ret *common.SlaveStatus, err *cd.Result) {
	statusFile := path.Join(serviceInfo.Volumes.BackPath.Value, "slave.json")

	ptr := &common.SlaveStatus{}
	fileErr := fu.LoadConfig(statusFile, ptr)
	if fileErr != nil {
		err = cd.NewError(cd.UnExpected, "load slave config failed")
		return
	}

	ret = ptr
	return
}

func (s *Base) saveSlaveStatus(serviceInfo *common.ServiceInfo, ptr *common.SlaveStatus) (err *cd.Result) {
	statusFile := path.Join(serviceInfo.Volumes.BackPath.Value, "slave.json")

	fileErr := fu.SaveConfig(statusFile, ptr)
	if fileErr != nil {
		err = cd.NewError(cd.UnExpected, "save slave config failed")
		return
	}
	return
}

func (s *Base) loadMasterStatus(serviceInfo *common.ServiceInfo) (ret *common.MasterStatus, err *cd.Result) {
	masterFile := path.Join(serviceInfo.Volumes.BackPath.Value, "master.json")

	ptr := &common.MasterStatus{}
	fileErr := fu.LoadConfig(masterFile, ptr)
	if fileErr != nil {
		err = cd.NewError(cd.UnExpected, "load master config failed")
		return
	}

	ret = ptr
	return
}

func (s *Base) saveMasterStatus(serviceInfo *common.ServiceInfo, ptr *common.MasterStatus) (err *cd.Result) {
	masterFile := path.Join(serviceInfo.Volumes.BackPath.Value, "master.json")

	fileErr := fu.SaveConfig(masterFile, ptr)
	if fileErr != nil {
		err = cd.NewError(cd.UnExpected, "save master config failed")
		return
	}
	return
}
