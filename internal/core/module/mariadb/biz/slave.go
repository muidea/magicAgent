package biz

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magieAgent/pkg/common"
)

func (s *Mariadb) ToSlave(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("ToSlave failed, nil param")
		return
	}

	serviceInfo, serviceOK := param.(*common.ServiceInfo)
	if !serviceOK {
		log.Warnf("ToSlave failed, illegal service name")
		return
	}

	masterInfo, masterOK := ev.GetData("master").(*common.MasterInfo)
	if !masterOK {
		log.Warnf("ToSlave %s failed, illegal master info", serviceInfo)
		return
	}

	log.Warnf("ToSlave, service:%s, master=[%v]", serviceInfo, masterInfo.Dump())

	// 如果已经处于slaver状态，则不用重复执行
	statusPtr, err := s.showSlaveStatus(serviceInfo)
	if err != nil {
		log.Errorf("ToSlave %s failed, s.showSlaveStatus error:%s", serviceInfo, err.Error())
		if re != nil {
			re.Set(nil, err)
		}

		return
	}
	// 当前已经处于slave状态，并且master没有发生变化
	if statusPtr.Enable && statusPtr.MasterHost == masterInfo.Host && statusPtr.MasterPort == masterInfo.Port {
		if re != nil {
			re.Set(nil, nil)
		}
		return
	}

	// 先主动停止，slave状态
	if statusPtr.Enable {
		err = s.stopSlave(serviceInfo)
		if err != nil {
			log.Errorf("ToSlave %s failed, s.stopSlave error:%s", serviceInfo, err.Error())
		}
	}

	err = s.changeMaster(serviceInfo, masterInfo)
	if err != nil {
		log.Errorf("ToSlave %s failed, s.changeMaster error:%s", serviceInfo, err.Error())
		if re != nil {
			re.Set(nil, err)
		}
		return
	}

	err = s.startSlave(serviceInfo)
	if err != nil {
		log.Errorf("ToSlave %s failed, s.startSlave error:%s", serviceInfo, err.Error())
	}

	if re != nil {
		re.Set(nil, err)
	}
}

func (s *Mariadb) CheckSlave(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("CheckSlave failed, nil param")
		return
	}

	serviceInfo, serviceOK := param.(*common.ServiceInfo)
	if !serviceOK {
		log.Warnf("CheckSlave failed, illegal service name")
		return
	}

	statusPtr, err := s.showSlaveStatus(serviceInfo)
	if err != nil {
		log.Errorf("CheckSlave %s failed, s.showSlaveStatus error:%s", serviceInfo, err.Error())
	}

	if re != nil {
		re.Set(statusPtr, err)
	}
}

func (s *Mariadb) Restore(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("Restore failed, nil param")
		return
	}

	serviceInfo, serviceOK := param.(*common.ServiceInfo)
	if !serviceOK {
		log.Warnf("Restore failed, illegal service name")
		return
	}

	// 1、先停止当前实例
	// 2、使用临时实例进行数据恢复
	// 3、恢复当前实例
	var err *cd.Result
	func() {
		err = s.stopService(serviceInfo)
		if err != nil {
			log.Errorf("Restore %s failed, s.stopService error:%s", serviceInfo, err.Error())
			return
		}

		err = s.restore(serviceInfo)
		_ = s.startService(serviceInfo)
		if err != nil {
			log.Errorf("Restore %s failed, s.restore error:%s", serviceInfo, err.Error())
			return
		}
	}()

	if err != nil {
		log.Errorf("Restore %v failed, error:%v", serviceInfo, err.Error())
	}

	if re != nil {
		re.Set(nil, err)
	}
}
