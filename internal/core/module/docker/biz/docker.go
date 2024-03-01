package biz

import (
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicAgent/pkg/common"
)

func (s *Docker) ExecuteCommand(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("ExecuteCommand failed, nil param")
		return
	}

	paramVal, paramOK := param.(*common.ServiceParam)
	if !paramOK {
		log.Warnf("ExecuteCommand failed, illegal param")
		return
	}

	cmd := []string{"sh", "-c", paramVal.CmdParam}
	cmdArgs := []string{"exec", paramVal.Service}
	cmdArgs = append(cmdArgs, cmd...)
	// 抽取返回值检查是否出错
	resultVal, errorVal, resultErr := s.Execute(cmdName, cmdArgs...)
	if re != nil {
		re.Set(resultVal, resultErr)
		re.SetVal("stderr", errorVal)
	}
}

func (s *Docker) StartService(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("StartService failed, nil param")
		return
	}

	paramVal, paramOK := param.(string)
	if !paramOK {
		log.Warnf("StartService failed, illegal param")
		return
	}

	cmdArgs := []string{"start", paramVal}
	// 抽取返回值检查是否出错
	resultVal, errorVal, resultErr := s.Execute(cmdName, cmdArgs...)
	if re != nil {
		re.Set(resultVal, resultErr)
		re.SetVal("stderr", errorVal)
	}
}

func (s *Docker) StopService(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("StopService failed, nil param")
		return
	}

	paramVal, paramOK := param.(string)
	if !paramOK {
		log.Warnf("StopService failed, illegal param")
		return
	}

	cmdArgs := []string{"stop", paramVal}
	// 抽取返回值检查是否出错
	resultVal, errorVal, resultErr := s.Execute(cmdName, cmdArgs...)
	if re != nil {
		re.Set(resultVal, resultErr)
		re.SetVal("stderr", errorVal)
	}
}
