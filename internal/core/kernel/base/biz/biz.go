package biz

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicCommon/foundation/net"
	"github.com/muidea/magicCommon/task"

	"github.com/muidea/magieAgent/internal/config"
	"github.com/muidea/magieAgent/internal/core/base/biz"
	"github.com/muidea/magieAgent/pkg/common"
)

type changeCollection struct {
	catalog2ServiceList common.Catalog2ServiceList
	configPtr           *config.CfgItem
	startTime           time.Time
	loopCount           int
}

type Base struct {
	biz.Base

	curConfigFileInfo os.FileInfo

	curConfigPtr *config.CfgItem

	curRedoCheckMutex sync.RWMutex
	curRedoCollection []*changeCollection

	currentKernelService sync.Map
}

func New(
	eventHub event.Hub,
	backgroundRoutine task.BackgroundRoutine,
) *Base {
	ptr := &Base{
		Base:              biz.New(common.BaseModule, eventHub, backgroundRoutine),
		curRedoCollection: []*changeCollection{},
	}

	ptr.SubscribeFunc(common.NotifyTimer, ptr.timerCheck)
	ptr.SubscribeFunc(common.NotifyService, ptr.serviceNotify)

	return ptr
}

func (s *Base) GetKernelHealthStatus() bool {
	statusVal := true
	s.currentKernelService.Range(func(_, v any) bool {
		serviceInfo := v.(*common.ServiceInfo)
		if serviceInfo.Status == nil || !serviceInfo.Status.Available {
			statusVal = false
		}

		return true
	})

	return statusVal
}

func (s *Base) GetKernelReadyStatus() bool {
	statusVal := true
	s.currentKernelService.Range(func(_, v any) bool {
		serviceInfo := v.(*common.ServiceInfo)
		_, slaveStatus, statusErr := s.QueryLocalMariadbStatus(serviceInfo.Name)
		if statusErr != nil {
			log.Warnf("GetKernelReadyStatus failed, s.QueryLocalMariadbStatus %v error:%v", serviceInfo, statusErr.Error())
			statusVal = false
			return true
		}

		if !slaveStatus.Enable {
			return true
		}
		if slaveStatus.BehindSecond > 0 {
			log.Warnf("GetKernelReadyStatus failed, s.QueryLocalMariadbStatus %v behind master:%d", serviceInfo, slaveStatus.BehindSecond)
			statusVal = false
		}
		return true
	})

	return statusVal
}

func (s *Base) addKernelService(serviceInfo *common.ServiceInfo) {
	if serviceInfo.Catalog != common.Mariadb {
		return
	}

	s.currentKernelService.Store(serviceInfo.Name, serviceInfo)
}

func (s *Base) delKernelService(serviceInfo *common.ServiceInfo) {
	if serviceInfo.Catalog != common.Mariadb {
		return
	}

	s.currentKernelService.Delete(serviceInfo.Name)
}

func (s *Base) timerCheck(_ event.Event, _ event.Result) {
	func() {
		fileInfo, fileErr := os.Stat(config.GetConfigFile())
		if fileErr != nil {
			return
		}

		if s.curConfigFileInfo != nil && fileInfo.ModTime() == s.curConfigFileInfo.ModTime() {
			return
		}

		curConfig := config.ReloadConfig()
		if curConfig.Status == nil || s.curConfigPtr == nil {
			s.curConfigFileInfo = fileInfo
			s.curConfigPtr = curConfig

			if s.curConfigPtr.Status != nil {
				s.statusChange(s.curConfigPtr)
			}
			return
		}

		// 如果状态未发生变化，则不用继续
		if !s.curConfigPtr.StatusChange(curConfig) {
			return
		}

		s.curConfigFileInfo = fileInfo
		s.curConfigPtr = curConfig
		s.statusChange(s.curConfigPtr)
	}()

	s.checkStatusRedo()
}

func (s *Base) serviceNotify(ev event.Event, _ event.Result) {
	serviceInfo := ev.Data().(*common.ServiceInfo)
	if serviceInfo == nil || s.curConfigPtr == nil {
		return
	}
	actionVal := ev.Header().Get(event.Action)
	if actionVal == event.Del {
		s.delKernelService(serviceInfo)
		return
	}

	s.addKernelService(serviceInfo)
	if serviceInfo.Status == nil || !serviceInfo.Status.Available {
		return
	}

	// 到这里说明当前service状态正常，这里主动的做一次状态切换
	// 同一个服务如果进行多次状态切换是允许的
	catalog2ServiceList := common.Catalog2ServiceList{
		serviceInfo.Catalog: common.ServiceList{serviceInfo.Name},
	}

	changePtr := &changeCollection{
		catalog2ServiceList: catalog2ServiceList,
		configPtr:           s.curConfigPtr.Dump(),
		startTime:           time.Now(),
		loopCount:           0,
	}

	s.curRedoCheckMutex.Lock()
	defer s.curRedoCheckMutex.Unlock()
	s.curRedoCollection = append(s.curRedoCollection, changePtr)
	return
}

func (s *Base) statusChange(cfgPtr *config.CfgItem) {
	if cfgPtr.Status == nil {
		log.Errorf("statusChange failed, status is nil")
		return
	}

	queryEvent := event.NewEvent(common.ListService, s.ID(), common.K8sModule, nil, common.Mariadb)
	result := s.SendEvent(queryEvent)
	resultVal, resultErr := result.Get()
	if resultErr != nil {
		log.Errorf("query service list failed, err:%s", resultErr.Error())
		return
	}

	catalog2ServiceList := resultVal.(common.Catalog2ServiceList)
	changePtr := &changeCollection{
		catalog2ServiceList: catalog2ServiceList,
		configPtr:           cfgPtr.Dump(),
		startTime:           time.Now(),
		loopCount:           0,
	}

	s.curRedoCheckMutex.Lock()
	defer s.curRedoCheckMutex.Unlock()
	s.curRedoCollection = append(s.curRedoCollection, changePtr)
	return
}

func (s *Base) checkStatusRedo() {
	var curList []*changeCollection
	func() {
		s.curRedoCheckMutex.Lock()
		s.curRedoCheckMutex.Unlock()
		curList = s.curRedoCollection
		s.curRedoCollection = []*changeCollection{}
	}()

	remainList := []*changeCollection{}
	for _, val := range curList {
		remainPtr := s.redoStatusChange(val)
		if remainPtr != nil {
			//if remainPtr.loopCount >= 10 {
			if time.Since(remainPtr.startTime) > 30*time.Minute {
				// 到这里说明执行超过30分钟都无法成功，不能继续处理了。
				log.Warnf("drop status change, startTime-%v, Status-%v, loop count:%v", remainPtr.startTime, remainPtr.configPtr.Status, remainPtr.loopCount)
				continue
			}
			remainList = append(remainList, remainPtr)
		}
	}

	if len(remainList) == 0 {
		return
	}

	s.curRedoCheckMutex.Lock()
	s.curRedoCheckMutex.Unlock()
	remainList = append(remainList, s.curRedoCollection...)
	s.curRedoCollection = remainList
}

func (s *Base) redoStatusChange(ptr *changeCollection) (ret *changeCollection) {
	failedCatalog2ServiceList := common.Catalog2ServiceList{}

	for key, val := range ptr.catalog2ServiceList {
		if key == common.Mariadb {
			serviceList := common.ServiceList{}
			for _, sv := range val {
				err := s.SwitchMariadb(sv, ptr.configPtr)
				if err != nil {
					serviceList = append(serviceList, sv)
				}
			}

			if len(serviceList) > 0 {
				failedCatalog2ServiceList[key] = serviceList
			}

			continue
		}
	}

	if len(failedCatalog2ServiceList) > 0 {
		ret = &changeCollection{
			catalog2ServiceList: failedCatalog2ServiceList,
			configPtr:           ptr.configPtr,
			startTime:           ptr.startTime,
			loopCount:           ptr.loopCount + 1,
		}
	}

	return
}

func (s *Base) queryLocalServiceInfo(serviceName, catalog string) (ret *common.ServiceInfo, err *cd.Result) {
	ev := event.NewEvent(common.QueryService, s.ID(), common.K8sModule, nil, serviceName)
	ev.SetData("catalog", catalog)
	result := s.SendEvent(ev)
	resultVal, resultErr := result.Get()
	if resultErr != nil {
		err = resultErr
		return
	}

	serviceInfoPtr, serviceInfoOK := resultVal.(*common.ServiceInfo)
	if !serviceInfoOK {
		err = cd.NewError(cd.UnExpected, "illegal service info")
		return
	}

	ret = serviceInfoPtr
	return
}

func (s *Base) queryRemoteServiceInfo(serviceName, catalog string) (ret *common.ServiceInfo, err *cd.Result) {
	result := &common.QueryServiceResult{}
	urlVal, _ := url.ParseRequestURI(fmt.Sprintf("http://%s:%s", config.GetRemoteHost(), config.GetNodePort()))
	urlVal.Path = strings.Join([]string{urlVal.Path, common.ApiVersion, common.QueryService}, "")

	param := &common.ServiceParam{
		Name:    serviceName,
		Catalog: catalog,
	}
	client := &http.Client{}
	defer client.CloseIdleConnections()
	_, httpErr := net.HTTPPost(client, urlVal.String(), param, result)
	if httpErr != nil {
		err = cd.NewError(cd.UnExpected, httpErr.Error())
		return
	}

	ret = result.ServiceInfo
	return
}

func (s *Base) changeServiceEndpoint(serviceInfo *common.ServiceInfo) (err *cd.Result) {
	remoteService, remoteErr := s.queryRemoteServiceInfo(serviceInfo.Name, serviceInfo.Catalog)
	if remoteErr != nil {
		err = remoteErr
		log.Errorf("changeServiceEndpoint failed, s.queryRemoteServiceInfo %v error:%s", serviceInfo, remoteErr.Error())
		return err
	}

	endpointPtr := &common.Endpoint{
		Host: remoteService.Svc.Host,
		Port: remoteService.Svc.Port,
	}

	ev := event.NewEvent(common.ChangeServiceEndpoints, s.ID(), common.K8sModule, nil, serviceInfo)
	ev.SetData("endpoint", endpointPtr)
	result := s.SendEvent(ev)
	err = result.Error()
	return
}

func (s *Base) restoreServiceEndpoint(serviceInfo *common.ServiceInfo) (err *cd.Result) {
	ev := event.NewEvent(common.RestoreServiceEndpoints, s.ID(), common.K8sModule, nil, serviceInfo)
	result := s.SendEvent(ev)
	err = result.Error()
	return
}
