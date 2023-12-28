package biz

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/log"
	fp "github.com/muidea/magicCommon/foundation/path"

	"github.com/muidea/magieAgent/internal/config"
	"github.com/muidea/magieAgent/internal/core/module/docker/pkg/mariadb"
	"github.com/muidea/magieAgent/pkg/common"
)

func (s *Docker) getDefaultServiceInfo(serviceName, catalog string) (ret *common.ServiceInfo) {
	ret = &common.ServiceInfo{
		Name:    serviceName,
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
				Value: path.Join(config.GetDataPath(), serviceName),
			},
			BackPath: &common.Path{
				Name:  "backPath",
				Type:  common.HostPath,
				Value: path.Join(config.GetBackPath(), serviceName),
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
	}

	ret.Labels["app"] = serviceName
	ret.Labels["catalog"] = catalog
	return
}

func (s *Docker) getYamlFile(cmdInfo *common.CmdInfo) (ret string, err *cd.Result) {
	yamlPath := path.Join(config.GetWorkspace(), cmdInfo.ServiceInfo.Catalog)
	pathErr := os.MkdirAll(yamlPath, 0750)
	if pathErr != nil {
		err = cd.NewError(cd.UnExpected, pathErr.Error())
		return
	}

	yamlName := fmt.Sprintf("%s.yml", cmdInfo.Service)
	ret = path.Join(yamlPath, yamlName)
	return
}

func (s *Docker) getCommandFile(cmdInfo *common.CmdInfo, dockerTemplate string) (ret string, err *cd.Result) {
	type CommandData struct {
		common.ServiceInfo
		Command string
	}

	yamlFile, yamlErr := s.getYamlFile(cmdInfo)
	if yamlErr != nil {
		err = yamlErr
		return
	}

	if fp.Exist(yamlFile) {
		ret = yamlFile
		return
	}

	byteBuffer := bytes.NewBufferString("")
	templateHandle, templateErr := template.New("service").Parse(dockerTemplate)
	if templateErr != nil {
		err = cd.NewError(cd.UnExpected, templateErr.Error())
		return
	}

	commandData := &CommandData{
		ServiceInfo: *cmdInfo.ServiceInfo,
		Command:     strings.Join(cmdInfo.Command, " && "),
	}
	templateErr = templateHandle.Execute(byteBuffer, commandData)
	if templateErr != nil {
		err = cd.NewError(cd.UnExpected, templateErr.Error())
		return
	}

	fileHandle, fileErr := os.OpenFile(yamlFile, os.O_CREATE|os.O_WRONLY, 0750)
	if fileErr != nil {
		err = cd.NewError(cd.UnExpected, fileErr.Error())
		return
	}
	defer fileHandle.Close()

	_, fileErr = byteBuffer.WriteTo(fileHandle)
	if fileErr != nil {
		err = cd.NewError(cd.UnExpected, fileErr.Error())
		return
	}

	ret = yamlFile
	return
}

func (s *Docker) loadCommandFile(yamlFile string) (ret []*common.ServiceInfo, err error) {
	fileHandle, fileErr := os.OpenFile(yamlFile, os.O_RDONLY, 0750)
	if fileErr != nil {
		err = cd.NewError(cd.UnExpected, fileErr.Error())
		return
	}
	defer fileHandle.Close()

	byteBuffer := bytes.NewBufferString("")
	_, fileErr = byteBuffer.ReadFrom(fileHandle)
	if fileErr != nil {
		err = cd.NewError(cd.UnExpected, fileErr.Error())
		return
	}

	var result interface{}
	fileErr = yaml.Unmarshal(byteBuffer.Bytes(), &result)
	if fileErr != nil {
		err = cd.NewError(cd.UnExpected, fileErr.Error())
		return
	}

	if config, ok := result.(map[interface{}]interface{}); ok {
		if services, ok := config["services"].(map[interface{}]interface{}); ok {
			for sk, sv := range services {
				if service, ok := sv.(map[interface{}]interface{}); ok {
					for ssk, ssv := range service {
						if ssk.(string) == "labels" {
							var serviceInfo *common.ServiceInfo
							if labels, ok := ssv.([]interface{}); ok {
								serviceInfo = &common.ServiceInfo{
									Name:    sk.(string),
									Volumes: &common.Volumes{},
									Env:     &common.Env{},
									Svc:     &common.Svc{},
								}
								for _, lv := range labels {
									if lvs, ok := lv.(string); ok {
										items := strings.Split(lvs, "=")
										if len(items) != 2 {
											continue
										}

										switch items[0] {
										case "service.catalog":
											serviceInfo.Catalog = items[1]
										case "service.image":
											serviceInfo.Image = items[1]
										case "service.confPath":
											serviceInfo.Volumes.ConfPath = &common.Path{
												Name:  "config",
												Type:  common.InnerPath,
												Value: items[1],
											}
										case "service.dataPath":
											serviceInfo.Volumes.DataPath = &common.Path{
												Name:  "dataPath",
												Type:  common.InnerPath,
												Value: items[1],
											}
										case "service.backPath":
											serviceInfo.Volumes.BackPath = &common.Path{
												Name:  "backPath",
												Type:  common.InnerPath,
												Value: items[1],
											}
										case "service.root":
											serviceInfo.Env.Root = items[1]
										case "service.password":
											serviceInfo.Env.Password = items[1]
										case "service.host":
											serviceInfo.Svc.Host = items[1]
										case "service.port":
											iVal, _ := strconv.ParseInt(items[1], 10, 64)
											serviceInfo.Svc.Port = int32(iVal)
										}
									}
								}
							}

							if serviceInfo != nil {
								ret = append(ret, serviceInfo)
							}
						}
					}
				}
			}
		}
	}

	return
}

func (s *Docker) ExecuteCommand(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("ExecuteCommand failed, nil param")
		return
	}

	cmdInfoPtr, cmdInfoOK := param.(*common.CmdInfo)
	if !cmdInfoOK || cmdInfoPtr == nil {
		log.Warnf("ExecuteCommand failed, illegal param")
		return
	}

	// docker exec mariadb001 /bin/bash -c 'mysql -uroot -prootkit -e "show master status\G;"'
	commandFile, commandErr := s.getCommandFile(cmdInfoPtr, mariadb.ServiceDockerTemplate)
	if commandErr != nil {
		if re != nil {
			re.Set(nil, commandErr)
		}
		return
	}

	cmdName := "docker-compose"
	cmdArgs := []string{"-f", commandFile, "exec", "-u", "root", cmdInfoPtr.Service}
	cmdArgs = append(cmdArgs, cmdInfoPtr.Command...)
	// 抽取返回值检查是否出错
	resultData, resultErr := s.Execute(cmdName, cmdArgs...)
	if re != nil {
		re.Set(resultData, resultErr)
	}
}

func (s *Docker) CreateService(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("CreateService failed, nil param")
		return
	}

	serviceName, serviceOK := param.(string)
	if !serviceOK {
		log.Warnf("CreateService failed, illegal param")
		return
	}
	catalog := ev.GetData("catalog")
	if catalog == nil {
		log.Warnf("CreateService failed, illegal catalog")
		return
	}
	catalogVal, catalogOK := catalog.(string)
	if !catalogOK || catalogVal != common.Mariadb {
		log.Warnf("CreateService failed, illegal catalog")
		return
	}

	err := s.Create(serviceName, catalogVal)
	if re != nil {
		re.Set(nil, err)
	}
}

func (s *Docker) DestroyService(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("DestroyService failed, nil param")
		return
	}

	serviceName, serviceOK := param.(string)
	if !serviceOK {
		log.Warnf("DestroyService failed, illegal param")
		return
	}
	catalog := ev.GetData("catalog")
	if catalog == nil {
		log.Warnf("DestroyService failed, illegal catalog")
		return
	}
	catalogVal, catalogOK := catalog.(string)
	if !catalogOK || catalogVal != common.Mariadb {
		log.Warnf("DestroyService failed, illegal catalog")
		return
	}

	err := s.Destroy(serviceName, catalogVal)
	if re != nil {
		re.Set(nil, err)
	}
}

func (s *Docker) StartService(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("StartService failed, nil param")
		return
	}

	cmdInfoPtr, cmdInfoOK := param.(*common.CmdInfo)
	if !cmdInfoOK {
		log.Warnf("StartService failed, illegal param")
		return
	}

	commandFile, commandErr := s.getCommandFile(cmdInfoPtr, mariadb.ServiceDockerTemplate)
	if commandErr != nil {
		if re != nil {
			re.Set(nil, commandErr)
		}
		return
	}

	cmdName := "docker-compose"
	cmdArgs := []string{"-f", commandFile, "start"}
	// 抽取返回值检查是否出错
	resultData, resultErr := s.Execute(cmdName, cmdArgs...)
	if re != nil {
		re.Set(resultData, resultErr)
	}
}

func (s *Docker) StopService(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("StopService failed, nil param")
		return
	}

	cmdInfoPtr, cmdInfoOK := param.(*common.CmdInfo)
	if !cmdInfoOK {
		log.Warnf("StopService failed, illegal param")
		return
	}

	commandFile, commandErr := s.getCommandFile(cmdInfoPtr, mariadb.ServiceDockerTemplate)
	if commandErr != nil {
		if re != nil {
			re.Set(nil, commandErr)
		}
		return
	}

	cmdName := "docker-compose"
	cmdArgs := []string{"-f", commandFile, "stop"}
	// 抽取返回值检查是否出错
	resultData, resultErr := s.Execute(cmdName, cmdArgs...)
	if re != nil {
		re.Set(resultData, resultErr)
	}
}

func (s *Docker) JobService(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("JobService failed, nil param")
		return
	}

	cmdInfoPtr, cmdInfoOK := param.(*common.CmdInfo)
	if !cmdInfoOK {
		log.Warnf("JobService failed, illegal param")
		return
	}

	commandFile, commandErr := s.getCommandFile(cmdInfoPtr, mariadb.JobDockerTemplate)
	if commandErr != nil {
		if re != nil {
			re.Set(nil, commandErr)
		}
		return
	}

	cmdName := "docker-compose"
	upCmdArgs := []string{"-f", commandFile, "up", "-d"}
	// 抽取返回值检查是否出错
	_, resultErr := s.Execute(cmdName, upCmdArgs...)
	if resultErr != nil {
		if re != nil {
			re.Set(nil, resultErr)
		}
		return
	}

	downCmdArgs := []string{"-f", commandFile, "down"}
	// 抽取返回值检查是否出错
	_, resultErr = s.Execute(cmdName, downCmdArgs...)
	if re != nil {
		re.Set(nil, resultErr)
	}
	return
}

func (s *Docker) ListService(ev event.Event, re event.Result) {
	var catalogList []string
	param := ev.Data()
	if param == nil {
		catalogList = common.DefaultCatalogList
	} else {
		catalogVal, catalogOK := param.(string)
		if !catalogOK {
			log.Warnf("ListService failed, illegal param")
			return
		}

		catalogList = append(catalogList, catalogVal)
	}

	catalog2ServiceList := s.enumService(catalogList)

	if re != nil {
		re.Set(catalog2ServiceList, nil)
	}
}

func (s *Docker) enumService(catalogList []string) common.Catalog2ServiceList {
	workspacePath := config.GetWorkspace()
	catalog2ServiceList := common.Catalog2ServiceList{}
	for _, ca := range catalogList {
		serviceList := common.ServiceList{}
		yamlPath := path.Join(workspacePath, ca)
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
				return nil
			}

			for _, val := range infoList {
				serviceList = append(serviceList, val.Name)
			}
			return nil
		})

		catalog2ServiceList[ca] = serviceList
	}

	return catalog2ServiceList
}

func (s *Docker) QueryService(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("QueryService failed, nil param")
		return
	}

	serviceName, serviceOK := param.(string)
	if !serviceOK {
		log.Warnf("QueryService failed, illegal param")
		return
	}

	catalog := ev.GetData("catalog")
	if catalog == nil || catalog.(string) != common.Mariadb {
		log.Warnf("QueryService failed, illegal catalog")
		return
	}

	serviceInfoPtr, serviceInfoErr := s.Query(serviceName, catalog.(string))
	if re != nil {
		re.Set(serviceInfoPtr, serviceInfoErr)
	}
}
