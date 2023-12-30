package base

import (
	engine "github.com/muidea/magicEngine"

	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/module"
	"github.com/muidea/magicCommon/task"

	"github.com/muidea/magicAgent/internal/core/kernel/base/biz"
	"github.com/muidea/magicAgent/internal/core/kernel/base/service"
	"github.com/muidea/magicAgent/pkg/common"
)

func init() {
	module.Register(New())
}

type Base struct {
	routeRegistry engine.Router

	service *service.Base
	biz     *biz.Base
}

func New() *Base {
	return &Base{}
}

func (s *Base) ID() string {
	return common.BaseModule
}

func (s *Base) BindRegistry(routeRegistry engine.Router) {

	s.routeRegistry = routeRegistry
}

func (s *Base) Setup(endpointName string, eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) {
	s.biz = biz.New(eventHub, backgroundRoutine)

	s.service = service.New(endpointName, s.biz)
	s.service.BindRegistry(s.routeRegistry)
	s.service.RegisterRoute()
}
