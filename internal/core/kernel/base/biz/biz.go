package biz

import (
	"fmt"
	"time"

	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicCommon/task"

	"github.com/muidea/magicAgent/internal/config"
	"github.com/muidea/magicAgent/internal/core/base/biz"
	"github.com/muidea/magicAgent/pkg/common"
)

type Base struct {
	biz.Base

	checkingFlag  bool
	unexpectCount int
	unexpectTime  time.Time
}

func New(
	eventHub event.Hub,
	backgroundRoutine task.BackgroundRoutine,
) *Base {
	ptr := &Base{
		Base: biz.New(common.BaseModule, eventHub, backgroundRoutine),
	}

	ptr.SubscribeFunc(common.NotifyTimer, ptr.timerCheck)

	return ptr
}

func (s *Base) timerCheck(_ event.Event, _ event.Result) {
	if s.checkingFlag {
		return
	}

	// 避免并发执行
	s.checkingFlag = true
	defer func() {
		s.checkingFlag = false
	}()

	mariadbService := config.GetGuards()
	for {
		currentTime := time.Now()
		statusPtr := s.queryMariadbStatus(mariadbService)
		if statusPtr != nil {
			unexpectFlag := true
			if statusPtr.IsNormal() {
				// 即使是当前节点状态正常，只有集群节点数量超过半数，才认为正常
				if statusPtr.NodeSize > len(config.GetClusterHosts())/2 {
					unexpectFlag = false
				}
			}

			if unexpectFlag {
				// 如果节点状态异常，则要进行异常计数
				if s.unexpectCount == 0 {
					s.unexpectTime = currentTime
				}

				s.unexpectCount++
			} else {
				if s.unexpectCount > 0 {
					log.Infof("Detected %s back to normal", mariadbService)
				}
				s.unexpectCount = 0
			}
		}

		if s.unexpectCount < 3 {
			break
		}

		// 持续超过3次检测异常，并且持续超过30s，这里就要考虑进行重启
		if time.Since(s.unexpectTime) < time.Duration(config.GetTimeOut())*time.Second {
			break
		}

		s.sendAlarmInfo(s.unexpectTime, mariadbService)
		// 一旦需要对节点进行重启，这里就要主动重置异常计数值
		s.restartMariadb(mariadbService)
		s.unexpectCount = 0
		break
	}

	return
}

func (s *Base) queryMariadbStatus(mariadbService string) *common.ClusterStatus {
	ev := event.NewEvent(common.QueryStatus, s.ID(), common.MariadbModule, nil, mariadbService)
	result := s.SendEvent(ev)
	statusVal, statusErr := result.Get()
	if statusErr != nil {
		if config.EnableTrace() {
			log.Errorf("queryMariadbStatus failed, error:%s", statusErr.Error())
		}
		return nil
	}

	return statusVal.(*common.ClusterStatus)
}

func (s *Base) restartMariadb(mariadbService string) {
	ev := event.NewEvent(common.StopService, s.ID(), common.DockerModule, nil, mariadbService)
	result := s.SendEvent(ev)
	_, stopErr := result.Get()
	if stopErr != nil {
		log.Errorf("restartMariadb failed, error:%s", stopErr.Error())
		return
	}

	ev = event.NewEvent(common.StartService, s.ID(), common.DockerModule, nil, mariadbService)
	result = s.SendEvent(ev)
	_, startErr := result.Get()
	if startErr != nil {
		log.Errorf("restartMariadb failed, error:%s", startErr.Error())
		return
	}
}

func (s *Base) sendAlarmInfo(timeStamp time.Time, mariadbService string) {
	alarmInfo := &common.AlarmInfo{
		Title: "Exception Alerts",
		Content: fmt.Sprintf("Node-%s service-%s exception was detected and a restart of the service is in progress. Exception time: %v, restart time: %v",
			config.GetLocalHost(),
			mariadbService,
			timeStamp,
			time.Now(),
		),
	}

	ev := event.NewEvent(common.SendAlarm, s.ID(), common.AlarmModule, nil, alarmInfo)
	s.PostEvent(ev)
}
