package common

import (
	"fmt"
	"time"

	cd "github.com/muidea/magicCommon/def"
)

const (
	NotifyRunning = "/running/notify/"
)

type ServiceParam struct {
	Name    string `json:"name"`
	Catalog string `json:"catalog"`
}

type CmdInfo struct {
	Service     string       `json:"service"`
	ServiceInfo *ServiceInfo `json:"serviceInfo"`
	Command     []string     `json:"command"`
}

type Labels map[string]string

func (s Labels) String() string {
	str := ""
	for k, v := range s {
		if str != "" {
			str = fmt.Sprintf("%s,%s=%s", str, k, v)
			continue
		}

		str = fmt.Sprintf("%s=%s", k, v)
	}
	return str
}

type Spec struct {
	CPU    string
	Memory string
}

type Path struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

type Volumes struct {
	ConfPath *Path `json:"confPath"`
	DataPath *Path `json:"dataPath"`
	BackPath *Path `json:"backPath"`
}

type Env struct {
	Root     string `json:"root"`
	Password string `json:"password"`
}

type Svc struct {
	Host string `json:"host"`
	Port int32  `json:"port"`
}

type Status struct {
	Available bool `json:"available"`
}

type ServiceInfo struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	Catalog   string   `json:"catalog"`
	Image     string   `json:"image"`
	Labels    Labels   `json:"labels"`
	Spec      *Spec    `json:"spec"`
	Volumes   *Volumes `json:"volumes"`
	Env       *Env     `json:"env"`
	Svc       *Svc     `json:"svc"`
	Replicas  int32    `json:"replicas"`
	Status    *Status  `json:"status"`
}

func (s *ServiceInfo) String() string {
	return fmt.Sprintf("%s:%s", s.Catalog, s.Name)
}

type ServiceList []string

type Catalog2ServiceList map[string]ServiceList

type TimerNotify struct {
	PreTime time.Time
	CurTime time.Time
}

type CreateServiceResult struct {
	cd.Result
}

type DestroyServiceResult struct {
	cd.Result
}

type StartServiceResult struct {
	cd.Result
}

type StopServiceResult struct {
	cd.Result
}

type QueryServiceResult struct {
	cd.Result
	ServiceInfo *ServiceInfo `json:"serviceInfo"`
}

type SyncServiceDataResult struct {
	cd.Result
}

type BackupMariadbDataResult struct {
	cd.Result
	DataPath string `json:"dataPath"`
}

type RestoreMariadbDataResult struct {
	cd.Result
}

const BaseModule = "/kernel/base"
