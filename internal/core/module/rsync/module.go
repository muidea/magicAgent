package rsync

import (
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/module"
	"github.com/muidea/magicCommon/task"

	engine "github.com/muidea/magicEngine"

	"github.com/muidea/magieAgent/internal/core/module/rsync/biz"
	"github.com/muidea/magieAgent/pkg/common"
)

func init() {
	module.Register(New())
}

type Rsync struct {
	routeRegistry engine.Router

	biz *biz.Rsync
}

func New() *Rsync {
	return &Rsync{}
}

func (s *Rsync) ID() string {
	return common.RsyncModule
}

func (s *Rsync) BindRegistry(routeRegistry engine.Router) {

	s.routeRegistry = routeRegistry
}

func (s *Rsync) Setup(endpointName string, eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) {
	s.biz = biz.New(eventHub, backgroundRoutine)
}
