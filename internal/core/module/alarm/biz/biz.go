package biz

import (
	"github.com/muidea/magicAgent/internal/config"
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicCommon/task"

	"github.com/muidea/magicAgent/internal/core/base/biz"
	"github.com/muidea/magicAgent/pkg/common"
)

type Alarm struct {
	biz.Base
}

func New(
	eventHub event.Hub,
	backgroundRoutine task.BackgroundRoutine,
) *Alarm {
	ptr := &Alarm{
		Base: biz.New(common.AlarmModule, eventHub, backgroundRoutine),
	}

	ptr.SubscribeFunc(common.SendAlarm, ptr.sendAlarm)

	return ptr
}

func (s *Alarm) sendAlarm(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("sendAlarm failed, nil param")
		return
	}

	err := s.SendAlarm(param.(*common.AlarmInfo))
	if re != nil {
		re.Set(nil, err)
	}
}

func (s *Alarm) SendAlarm(alarmInfo *common.AlarmInfo) (err *cd.Result) {
	emailServer := config.GetEMailInfo()
	if emailServer != nil {
		_ = s.sendEMail(alarmInfo, emailServer)
	}
	rayLinkServer := config.GetRayLinkInfo()
	if rayLinkServer != nil {
		_ = s.sendRayLink(alarmInfo, rayLinkServer)
	}

	return
}

func (s *Alarm) sendEMail(alarmInfo *common.AlarmInfo, emailServer *config.ServerInfo) (err *cd.Result) {
	return
}

func (s *Alarm) sendRayLink(alarmInfo *common.AlarmInfo, rayLink *config.ServerInfo) (err *cd.Result) {
	return
}
