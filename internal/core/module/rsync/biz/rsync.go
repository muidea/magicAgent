package biz

import (
	"fmt"

	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magieAgent/internal/config"
	"github.com/muidea/magieAgent/pkg/common"
)

func (s *Rsync) SyncFilesToRemote(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("SyncFilesToRemote failed, nil param")
		return
	}

	syncInfoPtr, syncInfoOK := param.(*common.SyncInfo)
	if !syncInfoOK || syncInfoPtr == nil {
		log.Warnf("SyncFilesToRemote failed, illegal param")
		return
	}

	source := fmt.Sprintf("%s/%s/", config.GetTmpPath(), syncInfoPtr.Local)
	destination := fmt.Sprintf("rsync://%s/%s/%s/", config.GetRemoteHost(), config.GetSyncSecret(), syncInfoPtr.Remote)

	cmdName := "rsync"
	args := []string{"-avz", "--password-file=/etc/rsyncd.passwd", source, destination}

	// 抽取返回值检查是否出错
	_, resultErr := s.Execute(cmdName, args...)
	if re != nil {
		re.Set(nil, resultErr)
	}
}

func (s *Rsync) SyncFilesToLocal(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("SyncFilesToRemote failed, nil param")
		return
	}

	syncInfoPtr, syncInfoOK := param.(*common.SyncInfo)
	if !syncInfoOK || syncInfoPtr == nil {
		log.Warnf("SyncFilesToRemote failed, illegal param")
		return
	}

	source := fmt.Sprintf("rsync://%s/%s/%s/", config.GetRemoteHost(), config.GetSyncSecret(), syncInfoPtr.Remote)
	destination := fmt.Sprintf("%s/%s/", config.GetTmpPath(), syncInfoPtr.Local)

	cmdName := "rsync"
	args := []string{"-avz", "--password-file=/etc/rsyncd.passwd", source, destination}

	// 抽取返回值检查是否出错
	_, resultErr := s.Execute(cmdName, args...)
	if re != nil {
		re.Set(nil, resultErr)
	}
}
