package config

import (
	"encoding/json"
	"os"

	fu "github.com/muidea/magicCommon/foundation/util"
)

var defaultConfig = `
{
    "localHost": "192.168.18.204",
    "clusterHosts": [
        "192.168.18.205"
    ],
    "guards": "mariadb001",
    "rayLink": {
        "serverUrl": "192.168.18.204",
        "account": "192.168.18.205",
        "password": "2023-11-15 18:20:00"
    },
    "email": {
        "serverUrl": "192.168.18.204",
        "account": "192.168.18.205",
        "password": "2023-11-15 18:20:00"
    }
}`

var currentWorkPath string
var configItem *CfgItem

const cfgPath = "/var/app/config/cfg.json"

func init() {
	err := LoadConfig(cfgPath)
	if err == nil {
		return
	}

	cfg := &CfgItem{}
	_ = json.Unmarshal([]byte(defaultConfig), cfg)
	configItem = cfg
	currentWorkPath, _ = os.Getwd()
}

func LoadConfig(cfgFile string) (err error) {
	if cfgFile == "" {
		return
	}

	cfg := &CfgItem{}
	err = fu.LoadConfig(cfgFile, cfg)
	if err != nil {
		return
	}

	configItem = cfg
	return
}

func GetWorkspace() string {
	return currentWorkPath
}

func GetLocalHost() string {
	return configItem.LocalHost
}

func GetClusterHosts() []string {
	return configItem.ClusterHosts
}

func GetGuards() string {
	return configItem.Guards
}

func GetTimeOut() int {
	return configItem.TimeOut
}

func GetRayLinkInfo() *ServerInfo {
	return configItem.RayLink
}

func GetEMailInfo() *ServerInfo {
	return configItem.EMail
}

type ServerInfo struct {
	ServerUrl string `json:"serverUrl"`
	Account   string `json:"account"`
	Password  string `json:"password"`
	Receiver  string `json:"receiver"`
}

type CfgItem struct {
	LocalHost    string      `json:"localHost"`
	ClusterHosts []string    `json:"clusterHosts"`
	Guards       string      `json:"guards"`
	TimeOut      int         `json:"timeOut"`
	RayLink      *ServerInfo `json:"rayLink"`
	EMail        *ServerInfo `json:"email"`
}
