package syncthing

import (
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/module"
	"github.com/muidea/magicCommon/task"

	engine "github.com/muidea/magicEngine"

	"github.com/muidea/magieAgent/internal/core/module/syncthing/biz"
	"github.com/muidea/magieAgent/pkg/common"
)

func init() {
	module.Register(New())
}

type Syncthing struct {
	routeRegistry engine.Router

	biz *biz.Syncthing
}

func New() *Syncthing {
	return &Syncthing{}
}

func (s *Syncthing) ID() string {
	return common.SyncthingModule
}

func (s *Syncthing) BindRegistry(routeRegistry engine.Router) {

	s.routeRegistry = routeRegistry
}

func (s *Syncthing) Setup(endpointName string, eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) {
	s.biz = biz.New(eventHub, backgroundRoutine)
}
