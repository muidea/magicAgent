package common

import cd "github.com/muidea/magicCommon/def"

const (
	SendAlarm = "/alarm/send"
)

type AlarmInfo struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type SendAlarmResult cd.Result

const AlarmModule = "/module/alarm"
