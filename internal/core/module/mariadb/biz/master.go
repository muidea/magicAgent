package biz

import (
	"fmt"
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magieAgent/pkg/common"
)

func (s *Mariadb) ToMaster(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("ToMaster failed, nil param")
		return
	}

	serviceInfo, serviceOK := param.(*common.ServiceInfo)
	if !serviceOK {
		log.Warnf("ToMaster failed, illegal service name")
		return
	}

	log.Warnf("ToMaster, service:%s", serviceInfo)

	grantErr := s.grantReplicate(serviceInfo)
	if grantErr != nil {
		log.Errorf("ToMaster %s failed, s.grantReplicate error:%s", serviceInfo, grantErr.Error())
		if re != nil {
			re.Set(nil, grantErr)
		}
		return
	}

	// 1、检查当前是否处于slave状态，并且存在同步延迟，如果有同步延迟，则不能切换，直接返回失败
	statusPtr, err := s.showSlaveStatus(serviceInfo)
	if err != nil {
		log.Errorf("ToMaster %s failed, s.showSlaveStatus error:%s", serviceInfo, err.Error())
		if re != nil {
			re.Set(nil, err)
		}

		return
	}

	if statusPtr.Enable && statusPtr.RunningOK && statusPtr.BehindSecond > 0 {
		log.Warnf("ToMaster %s failed, the current slave is %ds behind the master", serviceInfo, statusPtr.BehindSecond)
		if re != nil {
			err = cd.NewWarn(cd.Warned+1000, fmt.Sprintf("the current slave is %ds behind the master", statusPtr.BehindSecond))
			re.Set(nil, err)
		}
		return
	}

	// 如果当前处于slave状态还要主动停止一下
	if statusPtr.Enable {
		err = s.stopSlave(serviceInfo)
		if err != nil {
			log.Errorf("ToMaster %s failed, s.stopSlave error:%s", serviceInfo, err.Error())
		}

		if re != nil {
			re.Set(nil, err)
		}
		return
	}

	if re != nil {
		re.Set(nil, nil)
	}
}

func (s *Mariadb) CheckMaster(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("CheckMaster failed, nil param")
		return
	}

	serviceInfo, serviceOK := param.(*common.ServiceInfo)
	if !serviceOK {
		log.Warnf("CheckMaster failed, illegal service name")
		return
	}

	statusPtr, statusErr := s.showMasterStatus(serviceInfo)
	if statusErr != nil {
		log.Errorf("CheckMaster %s failed, s.showMasterStatus error:%s", serviceInfo, statusErr.Error())

		if re != nil {
			re.Set(nil, statusErr)
		}

		return
	}

	if re != nil {
		re.Set(statusPtr, statusErr)
	}
}

func (s *Mariadb) Backup(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("Backup failed, nil param")
		return
	}

	serviceInfo, serviceOK := param.(*common.ServiceInfo)
	if !serviceOK {
		log.Warnf("Backup failed, illegal service name")
		return
	}

	err := s.backup(serviceInfo)
	if err != nil {
		log.Errorf("Backup %s failed, s.backup error:%s", serviceInfo, err.Error())
		if re != nil {
			re.Set(nil, err)
		}
		return
	}

	if re != nil {
		dataPath := serviceInfo.Volumes.BackPath.Value
		re.Set(dataPath, err)
	}
	return
}
