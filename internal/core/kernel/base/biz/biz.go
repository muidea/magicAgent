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
	"github.com/muidea/magicCommon/foundation/net"
	"github.com/muidea/magicCommon/task"

	"github.com/muidea/magicAgent/internal/config"
	"github.com/muidea/magicAgent/internal/core/base/biz"
	"github.com/muidea/magicAgent/pkg/common"
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

	ptr.SubscribeFunc(common.NotifyService, ptr.serviceNotify)

	return ptr
}

func (s *Base) addKernelService(serviceInfo *common.ServiceInfo) {
	s.currentKernelService.Store(serviceInfo.Name, serviceInfo)
}

func (s *Base) delKernelService(serviceInfo *common.ServiceInfo) {
	s.currentKernelService.Delete(serviceInfo.Name)
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

func (s *Base) queryLocalServiceInfo(serviceName, catalog string) (ret *common.ServiceInfo, err *cd.Result) {
	ev := event.NewEvent(common.QueryService, s.ID(), common.DockerModule, nil, serviceName)
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
