package biz

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicCommon/task"

	"github.com/muidea/magieAgent/internal/core/base/biz"
	"github.com/muidea/magieAgent/pkg/common"
)

type Mariadb struct {
	biz.Base
}

func New(
	eventHub event.Hub,
	backgroundRoutine task.BackgroundRoutine,
) *Mariadb {
	ptr := &Mariadb{
		Base: biz.New(common.MariadbModule, eventHub, backgroundRoutine),
	}

	ptr.SubscribeFunc(common.SwitchToMasterStatus, ptr.ToMaster)
	ptr.SubscribeFunc(common.SwitchToSlaveStatus, ptr.ToSlave)
	ptr.SubscribeFunc(common.CheckMasterStatus, ptr.CheckMaster)
	ptr.SubscribeFunc(common.CheckSlaveStatus, ptr.CheckSlave)
	ptr.SubscribeFunc(common.BackupDatabase, ptr.Backup)
	ptr.SubscribeFunc(common.RestoreDatabase, ptr.Restore)

	return ptr
}

// grant replication slave on *.* to 'root'@'%' identified by 'rootkit';
func (s *Mariadb) grantReplicate(serviceInfo *common.ServiceInfo) (err *cd.Result) {
	cmd := fmt.Sprintf("grant replication slave on *.* to '%s'@'%%' identified by '%s';",
		serviceInfo.Env.Root,
		serviceInfo.Env.Password)

	log.Infof("grantReplicate, cmd:%s", cmd)
	_, err = s.execDatabaseCmd(serviceInfo, cmd)
	return
}

// change master to master_host='192.168.18.205', master_port=13306, master_user='root',master_password='rootkit', master_log_file='binlog.000002',master_log_pos=325
func (s *Mariadb) changeMaster(serviceInfo *common.ServiceInfo, masterInfo *common.MasterInfo) (err *cd.Result) {
	cmd := fmt.Sprintf("change master to master_host='%v', master_port=%d, master_user='%v',master_password='%v', master_log_file='%v',master_log_pos=%v;",
		masterInfo.Host,
		masterInfo.Port,
		serviceInfo.Env.Root,
		serviceInfo.Env.Password,
		masterInfo.LogFile,
		masterInfo.LogPos)

	log.Infof("changeMaster, cmd:%s", cmd)
	_, err = s.execDatabaseCmd(serviceInfo, cmd)
	return
}

// start slave
func (s *Mariadb) startSlave(serviceInfo *common.ServiceInfo) (err *cd.Result) {
	cmd := "start slave;"

	log.Infof("startSlave, cmd:%s", cmd)
	_, err = s.execDatabaseCmd(serviceInfo, cmd)
	return
}

// stop slave
func (s *Mariadb) stopSlave(serviceInfo *common.ServiceInfo) (err *cd.Result) {
	cmd := "stop slave;"

	log.Infof("stopSlave, cmd:%s", cmd)
	_, err = s.execDatabaseCmd(serviceInfo, cmd)
	return
}

var logFileExp = regexp.MustCompile(`File:\s*(\S+)`)

var logPosExp = regexp.MustCompile(`Position:\s*(\S+)`)

func (s *Mariadb) showMasterStatus(serviceInfo *common.ServiceInfo) (ret *common.MasterStatus, err *cd.Result) {
	cmd := "show master status\\G;"

	log.Infof("showMasterStatus, cmd:%s", cmd)
	statusVal, statusErr := s.execDatabaseCmd(serviceInfo, cmd)
	if statusErr != nil {
		err = statusErr
		return
	}

	var enable bool
	var logFile string
	var logPos int64
	scanner := bufio.NewScanner(bytes.NewReader(statusVal))
	for scanner.Scan() {
		items := logFileExp.FindStringSubmatch(scanner.Text())
		if len(items) > 1 {
			enable = true
			logFile = items[1]
			continue
		}

		items = logPosExp.FindStringSubmatch(scanner.Text())
		if len(items) > 1 {
			intVal, intErr := strconv.ParseInt(items[1], 10, 64)
			if intErr == nil {
				logPos = intVal
			}

			continue
		}
	}

	ret = &common.MasterStatus{
		Enable:    enable,
		Host:      serviceInfo.Svc.Host,
		Port:      serviceInfo.Svc.Port,
		LogFile:   logFile,
		LogPos:    logPos,
		TimeStamp: time.Now().UTC().UnixMilli(),
	}

	return
}

var slaveStatusExp = regexp.MustCompile(`Slave_IO_State:\s*(\S+)`)

var masterHostExp = regexp.MustCompile(`Master_Host:\s*(\S+)`)

var masterPortExp = regexp.MustCompile(`Master_Port:\s*(\S+)`)

var slaveIOExp = regexp.MustCompile(`Slave_IO_Running:\s*(\S+)`)

var slaveSQLExp = regexp.MustCompile(`Slave_SQL_Running:\s*(\S+)`)

var slaveBehindExp = regexp.MustCompile(`Seconds_Behind_Master:\s*(\S+)`)

func (s *Mariadb) showSlaveStatus(serviceInfo *common.ServiceInfo) (ret *common.SlaveStatus, err *cd.Result) {
	cmd := "show slave status\\G;"

	log.Infof("showSlaveStatus, cmd:%s", cmd)
	statusVal, statusErr := s.execDatabaseCmd(serviceInfo, cmd)
	if statusErr != nil {
		err = statusErr
		return
	}

	enableFlag := false
	masterHost := ""
	masterPort := int32(0)
	runningOK := false
	behindSecond := 0

	masterPortVal := ""
	slaveIOFlag := ""
	slaveSQLFlag := ""
	slaveBehindVal := ""
	scanner := bufio.NewScanner(bytes.NewReader(statusVal))
	for scanner.Scan() {
		items := slaveStatusExp.FindStringSubmatch(scanner.Text())
		if len(items) > 1 && items[1] != "" {
			enableFlag = true
			continue
		}

		items = masterHostExp.FindStringSubmatch(scanner.Text())
		if len(items) > 1 {
			masterHost = strings.ToLower(items[1])
			continue
		}

		items = masterPortExp.FindStringSubmatch(scanner.Text())
		if len(items) > 1 {
			masterPortVal = strings.ToLower(items[1])
			continue
		}

		items = slaveIOExp.FindStringSubmatch(scanner.Text())
		if len(items) > 1 {
			slaveIOFlag = strings.ToLower(items[1])
			continue
		}

		items = slaveSQLExp.FindStringSubmatch(scanner.Text())
		if len(items) > 1 {
			slaveSQLFlag = strings.ToLower(items[1])
			continue
		}

		items = slaveBehindExp.FindStringSubmatch(scanner.Text())
		if len(items) > 1 {
			slaveBehindVal = strings.ToLower(items[1])
			continue
		}
	}

	ret = &common.SlaveStatus{
		Enable:    enableFlag,
		TimeStamp: time.Now().UTC().UnixMilli(),
	}

	if !enableFlag {
		return
	}

	const yes = "yes"
	if (slaveIOFlag == slaveSQLFlag) && strings.Compare(slaveIOFlag, yes) == 0 {
		runningOK = true
	}
	intVal, intErr := strconv.ParseInt(slaveBehindVal, 10, 32)
	if intErr == nil {
		behindSecond = int(intVal)
	}
	intVal, intErr = strconv.ParseInt(masterPortVal, 10, 32)
	if intErr == nil {
		masterPort = int32(intVal)
	}

	ret.RunningOK = runningOK
	ret.MasterHost = masterHost
	ret.MasterPort = masterPort
	ret.BehindSecond = behindSecond

	return
}

func (s *Mariadb) execDatabaseCmd(serviceInfo *common.ServiceInfo, dbCmd string) (cmdResult []byte, err *cd.Result) {
	cmd := fmt.Sprintf("mysql -u%s -p%s -e", serviceInfo.Env.Root, serviceInfo.Env.Password)
	log.Infof("execDatabaseCmd %v, dbCmd:%s, args:%s", serviceInfo, dbCmd, cmd)

	cmdItems := strings.Split(cmd, " ")
	cmdItems = append(cmdItems, fmt.Sprintf("\"%s\"", dbCmd))
	cmdPtr := &common.CmdInfo{Service: serviceInfo.Name, ServiceInfo: serviceInfo, Command: cmdItems}
	ev := event.NewEvent(common.ExecuteCommand, s.ID(), common.K8sModule, nil, cmdPtr)
	result := s.SendEvent(ev)
	resultVal, resultErr := result.Get()
	if resultErr != nil {
		err = resultErr
		return
	}

	cmdResult = resultVal.([]byte)
	return
}

// mariabackup --defaults-file=/etc/mysql/conf.d/my.cnf --backup --rsync --host=192.168.18.204 --port=13306 --target-dir=/backup --datadir=/var/lib/mysql --user=root --password=rootkit
func (s *Mariadb) backup(serviceInfo *common.ServiceInfo) (err *cd.Result) {
	cleanCmd := "cd /backup/ && rm -rf * && ls /backup/"
	backupCmd := fmt.Sprintf("mariabackup --backup --host=%v --port=%v --target-dir=/backup --datadir=/var/lib/mysql --user=%v --password=%v",
		serviceInfo.Svc.Host,
		serviceInfo.Svc.Port,
		serviceInfo.Env.Root,
		serviceInfo.Env.Password,
	)
	prepareCmd := fmt.Sprintf("mariabackup --prepare --host=%v --port=%v --target-dir=/backup --datadir=/var/lib/mysql --user=%v --password=%v",
		serviceInfo.Svc.Host,
		serviceInfo.Svc.Port,
		serviceInfo.Env.Root,
		serviceInfo.Env.Password,
	)

	commands := []string{
		cleanCmd,
		backupCmd,
		prepareCmd,
	}
	log.Infof("backup %v, args:%s", serviceInfo, commands)
	cmdPtr := &common.CmdInfo{Service: fmt.Sprintf("%s-backup", serviceInfo.Name), ServiceInfo: serviceInfo, Command: commands}
	jobEvent := event.NewEvent(common.JobService, s.ID(), common.K8sModule, nil, cmdPtr)
	result := s.SendEvent(jobEvent)
	_, err = result.Get()
	if err != nil {
		return
	}

	return
}

// mariabackup --defaults-file=/etc/mysql/conf.d/my.cnf --copy-back --host=192.168.18.204 --port=13306 --target-dir=/backup --datadir=/var/lib/mysql --user=root --password=rootkit
func (s *Mariadb) restore(serviceInfo *common.ServiceInfo) (err *cd.Result) {
	cleanCmd := "cd /var/lib/mysql/ && rm -rf * && cd -"
	restoreCmd := fmt.Sprintf("mariabackup --copy-back --host=%v --port=%v --target-dir=/backup --datadir=/var/lib/mysql --user=%v --password=%v",
		serviceInfo.Svc.Host,
		serviceInfo.Svc.Port,
		serviceInfo.Env.Root,
		serviceInfo.Env.Password,
	)

	chownCmd := "chown -R mysql:mysql /var/lib/mysql/*"
	commands := []string{
		cleanCmd,
		restoreCmd,
		chownCmd,
	}

	log.Infof("restore %v, args:%s", serviceInfo, commands)
	cmdPtr := &common.CmdInfo{Service: fmt.Sprintf("%s-restore", serviceInfo.Name), ServiceInfo: serviceInfo, Command: commands}
	ev := event.NewEvent(common.JobService, s.ID(), common.K8sModule, nil, cmdPtr)
	result := s.SendEvent(ev)
	_, err = result.Get()

	log.Infof("restore %v finish!", serviceInfo)
	if err != nil {
		return
	}

	return
}

func (s *Mariadb) stopService(servicePtr *common.ServiceInfo) (err *cd.Result) {
	log.Infof("stopService %v", servicePtr)
	cmdPtr := &common.CmdInfo{Service: servicePtr.Name, ServiceInfo: servicePtr}
	ev := event.NewEvent(common.StopService, s.ID(), common.K8sModule, nil, cmdPtr)
	result := s.SendEvent(ev)
	err = result.Error()
	return
}

func (s *Mariadb) startService(servicePtr *common.ServiceInfo) (err *cd.Result) {
	log.Infof("startService %v", servicePtr)
	cmdPtr := &common.CmdInfo{Service: servicePtr.Name, ServiceInfo: servicePtr}
	ev := event.NewEvent(common.StartService, s.ID(), common.K8sModule, nil, cmdPtr)
	result := s.SendEvent(ev)
	err = result.Error()
	return
}
