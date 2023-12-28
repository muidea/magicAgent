package biz

import (
	"github.com/muidea/magieAgent/internal/config"
	"github.com/muidea/magieAgent/internal/core/module/docker/pkg/mariadb"
	"github.com/muidea/magieAgent/pkg/common"
	"path"
	"testing"
)

func TestGetCommandFile(t *testing.T) {
	dockerPtr := &Docker{}

	cmdInfo := &common.CmdInfo{
		Service: "t001",
		ServiceInfo: &common.ServiceInfo{
			Name:    "t001",
			Catalog: common.Mariadb,
			Image:   common.DefaultMariadbImage,
			Labels:  common.DefaultLabels,
			Volumes: &common.Volumes{
				ConfPath: &common.Path{
					Name:  "config",
					Type:  common.InnerPath,
					Value: path.Join(config.GetConfigPath(), "conf.d"),
				},
				DataPath: &common.Path{
					Name:  "dataPath",
					Type:  common.HostPath,
					Value: path.Join(config.GetDataPath(), "t001"),
				},
				BackPath: &common.Path{
					Name:  "backPath",
					Type:  common.HostPath,
					Value: path.Join(config.GetBackPath(), "t001"),
				},
			},
			Env: &common.Env{
				Root:     common.DefaultMariadbRoot,
				Password: common.DefaultMariadbPassword,
			},
			Svc: &common.Svc{
				Host: config.GetLocalHost(),
				Port: common.DefaultMariadbPort,
			},
			Replicas: 1,
		},
	}
	cmdFile, cmdErr := dockerPtr.getCommandFile(cmdInfo, mariadb.ServiceDockerTemplate)
	if cmdErr != nil {
		t.Errorf("getCommandFile failed, error:%s", cmdErr.Error())
		return
	}

	_, fileErr := dockerPtr.loadCommandFile(cmdFile)
	if fileErr != nil {
		t.Errorf("loadCommandFile failed, error:%s", fileErr.Error())
		return
	}

	t.Logf("cmdFile:%s", cmdFile)
}
