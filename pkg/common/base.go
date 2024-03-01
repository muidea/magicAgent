package common

import (
	"time"

	cd "github.com/muidea/magicCommon/def"
)

const (
	NotifyRunning = "/running/notify/"
)

type ServiceParam struct {
	Service  string `json:"service"`
	CmdParam string `json:"cmdParam"`
}

type TimerNotify struct {
	PreTime time.Time
	CurTime time.Time
}

type StartServiceResult struct {
	cd.Result
	StdOut string `json:"stdout"`
	StdErr string `json:"stderr"`
}

type StopServiceResult StartServiceResult

type ExecServiceResult StartServiceResult

const BaseModule = "/kernel/base"
