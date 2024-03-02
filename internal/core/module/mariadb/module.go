package docker

import (
	engine "github.com/muidea/magicEngine"

	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/module"
	"github.com/muidea/magicCommon/task"

	"github.com/muidea/magicAgent/internal/core/module/mariadb/biz"
	"github.com/muidea/magicAgent/internal/core/module/mariadb/service"
	"github.com/muidea/magicAgent/pkg/common"
)

func init() {
	module.Register(New())
}

type Mariadb struct {
	routeRegistry engine.Router

	service *service.Mariadb
	biz     *biz.Mariadb
}

func New() *Mariadb {
	return &Mariadb{}
}

func (s *Mariadb) ID() string {
	return common.MariadbModule
}

func (s *Mariadb) BindRegistry(routeRegistry engine.Router) {
	s.routeRegistry = routeRegistry
}

func (s *Mariadb) Setup(endpointName string, eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) {
	s.biz = biz.New(eventHub, backgroundRoutine)

	s.service = service.New(endpointName, s.biz)
	s.service.BindRegistry(s.routeRegistry)
	s.service.RegisterRoute()
}
