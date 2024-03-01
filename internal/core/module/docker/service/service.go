package service

import (
	"context"
	"net/http"

	engine "github.com/muidea/magicEngine"

	cd "github.com/muidea/magicCommon/def"
	fn "github.com/muidea/magicCommon/foundation/net"

	"github.com/muidea/magicAgent/internal/core/module/docker/biz"
	"github.com/muidea/magicAgent/pkg/common"
)

// Docker BaseService
type Docker struct {
	routeRegistry engine.Router

	bizPtr *biz.Docker

	endpointName string
}

// New create base
func New(endpointName string, bizPtr *biz.Docker) *Docker {
	ptr := &Docker{
		endpointName: endpointName,
		bizPtr:       bizPtr,
	}

	return ptr
}

func (s *Docker) BindRegistry(
	routeRegistry engine.Router) {

	s.routeRegistry = routeRegistry

	s.routeRegistry.SetApiVersion(common.ApiVersion)
}

// RegisterRoute 注册路由
func (s *Docker) RegisterRoute() {
	startRoute := engine.CreateRoute(common.StartService, engine.GET, s.StartHandle)
	s.routeRegistry.AddRoute(startRoute)

	stopRoute := engine.CreateRoute(common.StopService, engine.GET, s.StopHandle)
	s.routeRegistry.AddRoute(stopRoute)

	execRoute := engine.CreateRoute(common.ExecuteCommand, engine.POST, s.ExecHandle)
	s.routeRegistry.AddRoute(execRoute)
}

func (s *Docker) StartHandle(_ context.Context, res http.ResponseWriter, req *http.Request) {
	result := &common.StartServiceResult{}
	for {
		serviceName := req.URL.Query().Get("service")
		if serviceName == "" {
			result.ErrorCode = cd.IllegalParam
			result.Reason = "illegal service name"
			break
		}

		execStdout, execStderr, startErr := s.bizPtr.Start(serviceName)
		if startErr != nil {
			result.Result = *startErr
			break
		}

		result.StdOut = execStdout
		result.StdErr = execStderr
		break
	}

	fn.PackageHTTPResponse(res, result)
}

func (s *Docker) StopHandle(_ context.Context, res http.ResponseWriter, req *http.Request) {
	result := &common.StopServiceResult{}
	for {
		serviceName := req.URL.Query().Get("service")
		if serviceName == "" {
			result.ErrorCode = cd.IllegalParam
			result.Reason = "illegal service name"
			break
		}

		execStdout, execStderr, stopErr := s.bizPtr.Stop(serviceName)
		if stopErr != nil {
			result.Result = *stopErr
			break
		}

		result.StdOut = execStdout
		result.StdErr = execStderr
		break
	}

	fn.PackageHTTPResponse(res, result)
}

func (s *Docker) ExecHandle(_ context.Context, res http.ResponseWriter, req *http.Request) {
	result := &common.ExecServiceResult{}
	for {
		param := &common.ServiceParam{}
		err := fn.ParseJSONBody(req, nil, param)
		if err != nil {
			result.ErrorCode = cd.IllegalParam
			result.Reason = "非法参数"
			break
		}

		execStdout, execStderr, execErr := s.bizPtr.Exec(param.Service, param.CmdParam)
		if execErr != nil {
			result.Result = *execErr
			break
		}

		result.StdOut = execStdout
		result.StdErr = execStderr
		break
	}

	fn.PackageHTTPResponse(res, result)
}
