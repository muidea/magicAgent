package biz

import (
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"

	"github.com/muidea/magieAgent/internal/core/base/biz"
	"github.com/muidea/magieAgent/pkg/common"
)

type Rsync struct {
	biz.Base
}

func New(
	eventHub event.Hub,
	backgroundRoutine task.BackgroundRoutine,
) *Rsync {
	ptr := &Rsync{
		Base: biz.New(common.RsyncModule, eventHub, backgroundRoutine),
	}

	ptr.SubscribeFunc(common.SyncFilesToRemote, ptr.SyncFilesToRemote)

	ptr.SubscribeFunc(common.SyncFilesToLocal, ptr.SyncFilesToLocal)

	return ptr
}
