package biz

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"

	"github.com/muidea/magicAgent/internal/core/base/biz"
	"github.com/muidea/magicAgent/pkg/common"
)

const cmdName = "docker"

type Docker struct {
	biz.Base
}

func New(
	eventHub event.Hub,
	backgroundRoutine task.BackgroundRoutine,
) *Docker {
	ptr := &Docker{
		Base: biz.New(common.DockerModule, eventHub, backgroundRoutine),
	}

	ptr.SubscribeFunc(common.ExecuteCommand, ptr.ExecuteCommand)
	ptr.SubscribeFunc(common.StartService, ptr.StartService)
	ptr.SubscribeFunc(common.StopService, ptr.StopService)
	return ptr
}

func (s *Docker) Start(serviceName string) (stdout, stderr string, err *cd.Result) {
	cmdArgs := []string{"start", serviceName}
	// 抽取返回值检查是否出错
	resultVal, errorVal, resultErr := s.Execute(cmdName, cmdArgs...)
	if resultErr != nil {
		err = resultErr
		return
	}

	stdout = string(resultVal)
	stderr = string(errorVal)
	return
}

func (s *Docker) Stop(serviceName string) (stdout, stderr string, err *cd.Result) {
	cmdArgs := []string{"stop", serviceName}
	// 抽取返回值检查是否出错
	resultVal, errorVal, resultErr := s.Execute(cmdName, cmdArgs...)
	if resultErr != nil {
		err = resultErr
		return
	}

	stdout = string(resultVal)
	stderr = string(errorVal)
	return
}

func (s *Docker) Exec(serviceName, execParam string) (stdout, stderr string, err *cd.Result) {
	cmd := []string{"sh", "-c", execParam}
	cmdArgs := []string{"exec", serviceName}
	cmdArgs = append(cmdArgs, cmd...)
	// 抽取返回值检查是否出错
	resultVal, errorVal, resultErr := s.Execute(cmdName, cmdArgs...)
	if resultErr != nil {
		err = resultErr
		return
	}

	stdout = string(resultVal)
	stderr = string(errorVal)
	return
}
