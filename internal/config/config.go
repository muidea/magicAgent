package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	fu "github.com/muidea/magicCommon/foundation/util"
)

var defaultConfig = `
{
    "localHost": "192.168.18.204",
    "remoteHost": "192.168.18.205",
    "syncSecret": "PgDt4pxv6Uy9MMraAPjePSdZXnveYKtV",
    "dataPath": {
        "data": "/root/workspace/data"
    },
    "tmpPath": "/root/workspace/data/backup",
    "status": {
        "master": "192.168.18.204",
        "slave": "192.168.18.205",
        "updateTime": "2023-11-15 18:20:00"
    }
}`

var currentListenPort string
var currentNodePort string
var currentWorkPath string
var configItem *CfgItem

const cfgFile = "/var/local/share/cluster/cluster-info.json"

const (
	workspace = "workspace"
	config    = "config"
	data      = "data"
	backup    = "backup"
)

func init() {
	cfg := &CfgItem{}
	err := fu.LoadConfig(cfgFile, cfg)
	if err != nil {
		_ = json.Unmarshal([]byte(defaultConfig), cfg)
	}

	configItem = cfg

	currentWorkPath, _ = os.Getwd()
}

func SetListenPort(listenPort string) {
	currentListenPort = listenPort
}

func GetListenPort() string {
	return currentListenPort
}

func GetNodePort() string {
	nodePort, ok := os.LookupEnv("NODEPORT")
	if ok {
		return nodePort
	}

	return currentListenPort
}

func GetWorkspace() string {
	return path.Join(currentWorkPath, workspace)
}

func GetConfigPath() string {
	return path.Join(currentWorkPath, workspace, config)
}

func GetDataPath() string {
	return path.Join(currentWorkPath, workspace, data)
}

func GetBackPath() string {
	return path.Join(currentWorkPath, workspace, data, backup)
}

func GetConfigFile() string {
	return cfgFile
}

func ReloadConfig() *CfgItem {
	cfg := &CfgItem{}
	err := fu.LoadConfig(cfgFile, cfg)
	if err != nil {
		return nil
	}

	configItem = cfg
	return configItem
}

func GetLocalHost() string {
	return configItem.LocalHost
}

func GetRemoteHost() string {
	return configItem.RemoteHost
}

func GetSyncSecret() string {
	return configItem.SyncSecret
}

func GetTmpPath() string {
	return strings.TrimRight(configItem.TmpPath, "/")
}

type DataPath struct {
	Data string `json:"data"`
}

type Status struct {
	Master     string `json:"master"`
	Slave      string `json:"slave"`
	UpdateTime string `json:"updateTime"`
}

func (s *Status) IsSame(ptr *Status) bool {
	return s.Master == ptr.Master && s.Slave == ptr.Slave
}

func (s *Status) String() string {
	return fmt.Sprintf("Master:%s, Slave:%s, updateTime:%v", s.Master, s.Slave, s.UpdateTime)
}

type CfgItem struct {
	LocalHost  string    `json:"localHost"`
	RemoteHost string    `json:"remoteHost"`
	SyncSecret string    `json:"syncSecret"`
	DataPath   *DataPath `json:"dataPath"`
	TmpPath    string    `json:"tmpPath"`
	Status     *Status   `json:"status"`
}

func (s *CfgItem) StatusChange(cfgPtr *CfgItem) bool {
	if s.Status == nil && cfgPtr.Status != nil {
		return true
	}

	if s.Status != nil && cfgPtr.Status == nil {
		return true
	}

	return !s.Status.IsSame(cfgPtr.Status)
}

func (s *CfgItem) Dump() *CfgItem {
	ptr := &CfgItem{
		LocalHost:  s.LocalHost,
		RemoteHost: s.RemoteHost,
		SyncSecret: s.SyncSecret,
		TmpPath:    s.TmpPath,
	}

	if s.DataPath != nil {
		ptr.DataPath = &DataPath{Data: s.DataPath.Data}
	}
	if s.Status != nil {
		ptr.Status = &Status{
			Master:     s.Status.Master,
			Slave:      s.Status.Slave,
			UpdateTime: s.Status.UpdateTime,
		}
	}

	return ptr
}
