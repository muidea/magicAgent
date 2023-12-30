package biz

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicCommon/task"

	"github.com/muidea/magicAgent/internal/config"
	"github.com/muidea/magicAgent/internal/core/base/biz"
	"github.com/muidea/magicAgent/pkg/common"
)

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
	ptr.SubscribeFunc(common.JobService, ptr.JobService)
	ptr.SubscribeFunc(common.ListService, ptr.ListService)
	ptr.SubscribeFunc(common.QueryService, ptr.QueryService)
	ptr.SubscribeFunc(common.CreateService, ptr.CreateService)
	ptr.SubscribeFunc(common.DestroyService, ptr.DestroyService)
	return ptr
}

func (s *Docker) Create(serviceName, catalog string) (err *cd.Result) {
	cmdInfoPtr := &common.CmdInfo{
		Service: serviceName,
	}
	commandFile, commandErr := s.getCommandFile(cmdInfoPtr, "mariadb.ServiceDockerTemplate")
	if commandErr != nil {
		err = commandErr
		return
	}

	cmdName := "docker-compose"
	cmdArgs := []string{"-f", commandFile, "up", "-d"}
	// 抽取返回值检查是否出错
	_, resultErr := s.Execute(cmdName, cmdArgs...)
	if resultErr != nil {
		err = resultErr
	}
	return
}

func (s *Docker) Destroy(serviceName, catalog string) (err *cd.Result) {
	cmdInfoPtr := &common.CmdInfo{
		Service: serviceName,
	}
	commandFile, commandErr := s.getCommandFile(cmdInfoPtr, "mariadb.ServiceDockerTemplate")
	if commandErr != nil {
		err = commandErr
		return
	}

	cmdName := "docker-compose"
	cmdArgs := []string{"-f", commandFile, "down"}
	// 抽取返回值检查是否出错
	_, resultErr := s.Execute(cmdName, cmdArgs...)
	if resultErr != nil {
		err = resultErr
	}

	_ = os.Remove(commandFile)
	return
}

func (s *Docker) Start(serviceName, catalog string) (err *cd.Result) {
	cmdInfoPtr := &common.CmdInfo{
		Service: serviceName,
	}
	commandFile, commandErr := s.getCommandFile(cmdInfoPtr, "mariadb.ServiceDockerTemplate")
	if commandErr != nil {
		err = commandErr
		return
	}

	cmdName := "docker-compose"
	cmdArgs := []string{"-f", commandFile, "start"}
	// 抽取返回值检查是否出错
	_, resultErr := s.Execute(cmdName, cmdArgs...)
	if resultErr != nil {
		err = resultErr
	}
	return
}

func (s *Docker) Stop(serviceName, catalog string) (err *cd.Result) {
	cmdInfoPtr := &common.CmdInfo{
		Service: serviceName,
	}
	commandFile, commandErr := s.getCommandFile(cmdInfoPtr, "mariadb.ServiceDockerTemplate")
	if commandErr != nil {
		err = commandErr
		return
	}

	cmdName := "docker-compose"
	cmdArgs := []string{"-f", commandFile, "stop"}
	// 抽取返回值检查是否出错
	_, resultErr := s.Execute(cmdName, cmdArgs...)
	if resultErr != nil {
		err = resultErr
	}
	return
}

func (s *Docker) Query(serviceName, catalog string) (ret *common.ServiceInfo, err *cd.Result) {
	var resultErr *cd.Result
	var serviceInfoPtr *common.ServiceInfo
	workspacePath := config.GetWorkspace()
	yamlPath := path.Join(workspacePath, catalog)
	filepath.Walk(yamlPath, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}

		infoList, infoErr := s.loadCommandFile(path)
		if infoErr != nil {
			log.Errorf("s.loadCommandFile error:%s", infoErr.Error())
			resultErr = cd.NewError(cd.UnExpected, infoErr.Error())
			return nil
		}

		for _, val := range infoList {
			if val.Name == serviceName {
				serviceInfoPtr = val
				break
			}
		}

		return nil
	})

	if serviceInfoPtr == nil {
		resultErr = cd.NewError(cd.UnExpected, fmt.Sprintf("%s not exist", serviceName))
	}

	if resultErr != nil {
		err = resultErr
		return
	}

	ret = serviceInfoPtr
	return
}
