package pkg

import "fmt"

const (
	SwitchToMasterStatus = "/status/master/switch"
	SwitchToSlaveStatus  = "/status/slave/switch"
	CheckMasterStatus    = "/status/master/check"
	CheckSlaveStatus     = "/status/slave/check"
	BackupDatabase       = "/database/backup"
	RestoreDatabase      = "/database/restore"
)

const (
	DefaultMariadbImage      = "registry.supos.ai/jenkins/mariadb:10.6.11"
	DefaultMariadbRoot       = "root"
	DefaultMariadbPassword   = "rootkit"
	DefaultMariadbConfigPath = "/etc/conf.d"
	DefaultMariadbDataPath   = "/var/lib/mysql"
	DefaultMariadbBackPath   = "/backup"
	DefaultMariadbPort       = 3306
)

type MasterInfo struct {
	Host    string `json:"host"`
	Port    int32  `json:"port"`
	LogFile string `json:"logFile"`
	LogPos  int64  `json:"logPos"`
}

func (s *MasterInfo) Dump() string {
	return fmt.Sprintf("Host:%v, Port:%v, LogFile:%v, LogPos:%v", s.Host, s.Port, s.LogFile, s.LogPos)
}

type MasterStatus struct {
	Enable    bool   `json:"enable"`
	Host      string `json:"host"`
	Port      int32  `json:"port"`
	LogFile   string `json:"logFile"`
	LogPos    int64  `json:"logPos"`
	TimeStamp int64  `json:"timeStamp"`
}

func (s *MasterStatus) IsOK() bool {
	return s.LogFile != "" && s.LogPos > 0
}

type SlaveStatus struct {
	Enable       bool   `json:"enable"`
	MasterHost   string `json:"masterHost"`
	MasterPort   int32  `json:"masterPort"`
	RunningOK    bool   `json:"runningOK"`
	BehindSecond int    `json:"behindSecond"`
	TimeStamp    int64  `json:"timeStamp"`
}

func (s *SlaveStatus) IsOK() bool {
	if !s.Enable {
		return true
	}

	return s.RunningOK && s.BehindSecond > 0
}

const MariadbModule = "/module/mariadb"
