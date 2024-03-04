package docker

import (
	engine "github.com/muidea/magicEngine"

	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/module"
	"github.com/muidea/magicCommon/task"

	"github.com/muidea/magicAgent/internal/core/module/alarm/biz"
	"github.com/muidea/magicAgent/internal/core/module/alarm/service"
	"github.com/muidea/magicAgent/pkg/common"
)

func init() {
	module.Register(New())
}

type Alarm struct {
	routeRegistry engine.Router

	service *service.Alarm
	biz     *biz.Alarm
}

func New() *Alarm {
	return &Alarm{}
}

func (s *Alarm) ID() string {
	return common.AlarmModule
}

func (s *Alarm) BindRegistry(routeRegistry engine.Router) {
	s.routeRegistry = routeRegistry
}

func (s *Alarm) Setup(endpointName string, eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) {
	s.biz = biz.New(eventHub, backgroundRoutine)

	s.service = service.New(endpointName, s.biz)
	s.service.BindRegistry(s.routeRegistry)
	s.service.RegisterRoute()
}
