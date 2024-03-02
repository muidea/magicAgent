package config

import (
	"encoding/json"
	"os"

	fu "github.com/muidea/magicCommon/foundation/util"
)

var defaultConfig = `
{
    "localHost": "192.168.18.204",
    "remoteHost": [
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

const cfgFile = "/var/app/config/cfg.json"

func init() {
	cfg := &CfgItem{}
	err := fu.LoadConfig(cfgFile, cfg)
	if err != nil {
		_ = json.Unmarshal([]byte(defaultConfig), cfg)
	}

	configItem = cfg

	currentWorkPath, _ = os.Getwd()
}

func GetWorkspace() string {
	return currentWorkPath
}

func GetLocalHost() string {
	return configItem.LocalHost
}

func GetRemoteHost() []string {
	return configItem.RemoteHost
}

func GetGuards() string {
	return configItem.Guards
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
}

type CfgItem struct {
	LocalHost  string      `json:"localHost"`
	RemoteHost []string    `json:"remoteHost"`
	Guards     string      `json:"guards"`
	RayLink    *ServerInfo `json:"rayLink"`
	EMail      *ServerInfo `json:"email"`
}
