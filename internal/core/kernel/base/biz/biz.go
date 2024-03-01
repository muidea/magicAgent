package biz

import (
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"

	"github.com/muidea/magicAgent/internal/core/base/biz"
	"github.com/muidea/magicAgent/pkg/common"
)

type Base struct {
	biz.Base
}

func New(
	eventHub event.Hub,
	backgroundRoutine task.BackgroundRoutine,
) *Base {
	ptr := &Base{
		Base: biz.New(common.BaseModule, eventHub, backgroundRoutine),
	}

	ptr.SubscribeFunc(common.NotifyService, ptr.serviceNotify)

	return ptr
}

func (s *Base) serviceNotify(ev event.Event, _ event.Result) {
	return
}
