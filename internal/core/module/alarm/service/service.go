package service

import (
	"context"
	"net/http"

	cd "github.com/muidea/magicCommon/def"
	fn "github.com/muidea/magicCommon/foundation/net"

	engine "github.com/muidea/magicEngine"

	"github.com/muidea/magicAgent/internal/core/module/alarm/biz"
	"github.com/muidea/magicAgent/pkg/common"
)

// Alarm BaseService
type Alarm struct {
	routeRegistry engine.Router

	bizPtr *biz.Alarm

	endpointName string
}

// New create base
func New(endpointName string, bizPtr *biz.Alarm) *Alarm {
	ptr := &Alarm{
		endpointName: endpointName,
		bizPtr:       bizPtr,
	}

	return ptr
}

func (s *Alarm) BindRegistry(
	routeRegistry engine.Router) {

	s.routeRegistry = routeRegistry

	s.routeRegistry.SetApiVersion(common.ApiVersion)
}

// RegisterRoute 注册路由
func (s *Alarm) RegisterRoute() {
	statusRoute := engine.CreateRoute(common.SendAlarm, engine.POST, s.SendAlarmHandle)
	s.routeRegistry.AddRoute(statusRoute)
}

func (s *Alarm) SendAlarmHandle(_ context.Context, res http.ResponseWriter, req *http.Request) {
	result := &common.SendAlarmResult{}
	for {
		param := &common.AlarmInfo{}
		err := fn.ParseJSONBody(req, nil, param)
		if err != nil {
			result.ErrorCode = cd.IllegalParam
			result.Reason = "非法参数"
			break
		}

		sendErr := s.bizPtr.SendAlarm(param)
		if sendErr != nil {
			result.ErrorCode = sendErr.ErrorCode
			result.Reason = sendErr.Reason
		}
		break
	}

	fn.PackageHTTPResponse(res, result)
}
