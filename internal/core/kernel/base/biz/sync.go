package biz

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magieAgent/pkg/common"
)

func (s *Base) SyncDataToRemote(serviceName, catalog string, syncInfo *common.SyncInfo) (err *cd.Result) {
	_, serviceErr := s.queryLocalServiceInfo(serviceName, catalog)
	if serviceErr != nil {
		log.Errorf("SyncDataToRemote failed, s.queryLocalServiceInfo service:%s error:%s", serviceName, serviceErr.Error())
		err = serviceErr
		return
	}

	syncEvent := event.NewEvent(common.SyncFilesToRemote, s.ID(), common.SyncthingModule, nil, syncInfo)
	result := s.SendEvent(syncEvent)
	err = result.Error()
	return
}

func (s *Base) SyncDataToLocal(serviceName, catalog string, syncInfo *common.SyncInfo) (err *cd.Result) {
	_, serviceErr := s.queryLocalServiceInfo(serviceName, catalog)
	if serviceErr != nil {
		log.Errorf("SyncDataToLocal failed, s.queryLocalServiceInfo service:%s error:%s", serviceName, serviceErr.Error())
		err = serviceErr
		return
	}

	syncEvent := event.NewEvent(common.SyncFilesToLocal, s.ID(), common.SyncthingModule, nil, syncInfo)
	result := s.SendEvent(syncEvent)
	err = result.Error()
	return
}
