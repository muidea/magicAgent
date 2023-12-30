package docker

import (
	engine "github.com/muidea/magicEngine"

	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/module"
	"github.com/muidea/magicCommon/task"

	"github.com/muidea/magicAgent/internal/core/module/docker/biz"
	"github.com/muidea/magicAgent/internal/core/module/docker/service"
	"github.com/muidea/magicAgent/pkg/common"
)

func init() {
	module.Register(New())
}

type Docker struct {
	routeRegistry engine.Router

	service *service.Docker
	biz     *biz.Docker
}

func New() *Docker {
	return &Docker{}
}

func (s *Docker) ID() string {
	return common.DockerModule
}

func (s *Docker) BindRegistry(routeRegistry engine.Router) {

	s.routeRegistry = routeRegistry
}

func (s *Docker) Setup(endpointName string, eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) {
	s.biz = biz.New(eventHub, backgroundRoutine)

	s.service = service.New(endpointName, s.biz)
	s.service.BindRegistry(s.routeRegistry)
	s.service.RegisterRoute()
}
