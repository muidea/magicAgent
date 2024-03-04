package biz

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicCommon/foundation/net"
	"github.com/muidea/magicCommon/task"

	"github.com/muidea/magicAgent/internal/config"
	"github.com/muidea/magicAgent/internal/core/base/biz"
	"github.com/muidea/magicAgent/pkg/common"
)

type Alarm struct {
	biz.Base
}

func New(
	eventHub event.Hub,
	backgroundRoutine task.BackgroundRoutine,
) *Alarm {
	ptr := &Alarm{
		Base: biz.New(common.AlarmModule, eventHub, backgroundRoutine),
	}

	ptr.SubscribeFunc(common.SendAlarm, ptr.sendAlarm)

	return ptr
}

func (s *Alarm) sendAlarm(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("sendAlarm failed, nil param")
		return
	}

	err := s.SendAlarm(param.(*common.AlarmInfo))
	if re != nil {
		re.Set(nil, err)
	}
}

func (s *Alarm) SendAlarm(alarmInfo *common.AlarmInfo) (err *cd.Result) {
	emailServer := config.GetEMailInfo()
	if emailServer != nil {
		_ = s.sendEMail(alarmInfo, emailServer)
	}
	rayLinkServer := config.GetRayLinkInfo()
	if rayLinkServer != nil {
		_ = s.sendRayLink(alarmInfo, rayLinkServer)
	}

	return
}

func (s *Alarm) sendEMail(alarmInfo *common.AlarmInfo, emailServer *config.ServerInfo) (err *cd.Result) {
	sendErr := net.SendMail(
		emailServer.Account,
		emailServer.Password,
		emailServer.ServerUrl,
		[]string{emailServer.Receiver},
		alarmInfo.Title,
		alarmInfo.Content,
		[]string{},
		"text",
	)
	if sendErr == nil {
		return
	}

	log.Errorf("sendEMail failed, error:%s", sendErr.Error())
	err = cd.NewError(cd.UnExpected, sendErr.Error())
	return
}

/*
{
    "toJobNos": "20070111",
    "type": "oa",
    "body": {
        "title": "SPC异常消息通知2222",
        "content": "测试"
    }
}

serverUrl: http://10.192.20.6:50000/RESTAdapter/ALL/sendMsgByZLSPCToMSB
user: zlmes_pro
password: qwert123
*/

type RayLinkMessage struct {
	ToJobNos string            `json:"toJobNos"`
	Type     string            `json:"type"`
	Body     *common.AlarmInfo `json:"body"`
}

func (s *Alarm) sendRayLink(alarmInfo *common.AlarmInfo, rayLink *config.ServerInfo) (err *cd.Result) {
	msg := &RayLinkMessage{
		ToJobNos: rayLink.Receiver,
		Type:     "oa",
		Body:     alarmInfo,
	}

	header := url.Values{}
	header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(rayLink.Account+":"+rayLink.Password))))
	httpClient := &http.Client{}
	responseVal, responseErr := net.HTTPPost(httpClient, rayLink.ServerUrl, msg, nil, header)
	if responseErr == nil {
		return
	}

	log.Errorf("sendRayLink failed, error:%s, response:%s", responseErr.Error(), responseVal)
	err = cd.NewError(cd.UnExpected, responseErr.Error())
	return
}
