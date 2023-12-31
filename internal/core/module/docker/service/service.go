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
	createRoute := engine.CreateRoute(common.CreateService, engine.POST, s.CreateHandle)
	s.routeRegistry.AddRoute(createRoute)

	destroyRoute := engine.CreateRoute(common.DestroyService, engine.POST, s.DestroyHandle)
	s.routeRegistry.AddRoute(destroyRoute)

	startRoute := engine.CreateRoute(common.StartService, engine.POST, s.StartHandle)
	s.routeRegistry.AddRoute(startRoute)

	stopRoute := engine.CreateRoute(common.StopService, engine.POST, s.StopHandle)
	s.routeRegistry.AddRoute(stopRoute)

	queryRoute := engine.CreateRoute(common.QueryService, engine.POST, s.QueryHandle)
	s.routeRegistry.AddRoute(queryRoute)
}

func (s *Docker) CreateHandle(_ context.Context, res http.ResponseWriter, req *http.Request) {
	result := &common.CreateServiceResult{}
	for {
		param := &common.ServiceParam{}
		err := fn.ParseJSONBody(req, nil, param)
		if err != nil {
			result.ErrorCode = cd.IllegalParam
			result.Reason = "非法参数"
			break
		}
		createErr := s.bizPtr.Create(param.Name, param.Catalog)
		if createErr != nil {
			result.Result = *createErr
			break
		}

		break
	}

	fn.PackageHTTPResponse(res, result)
}

func (s *Docker) DestroyHandle(_ context.Context, res http.ResponseWriter, req *http.Request) {
	result := &common.DestroyServiceResult{}
	for {
		param := &common.ServiceParam{}
		err := fn.ParseJSONBody(req, nil, param)
		if err != nil {
			result.ErrorCode = cd.IllegalParam
			result.Reason = "非法参数"
			break
		}
		destroyErr := s.bizPtr.Destroy(param.Name, param.Catalog)
		if destroyErr != nil {
			result.Result = *destroyErr
			break
		}
		break
	}

	fn.PackageHTTPResponse(res, result)
}

func (s *Docker) StartHandle(_ context.Context, res http.ResponseWriter, req *http.Request) {
	result := &common.StartServiceResult{}
	for {
		param := &common.ServiceParam{}
		err := fn.ParseJSONBody(req, nil, param)
		if err != nil {
			result.ErrorCode = cd.IllegalParam
			result.Reason = "非法参数"
			break
		}
		startErr := s.bizPtr.Start(param.Name, param.Catalog)
		if startErr != nil {
			result.Result = *startErr
			break
		}

		break
	}

	fn.PackageHTTPResponse(res, result)
}

func (s *Docker) StopHandle(_ context.Context, res http.ResponseWriter, req *http.Request) {
	result := &common.StopServiceResult{}
	for {
		param := &common.ServiceParam{}
		err := fn.ParseJSONBody(req, nil, param)
		if err != nil {
			result.ErrorCode = cd.IllegalParam
			result.Reason = "非法参数"
			break
		}
		stopErr := s.bizPtr.Stop(param.Name, param.Catalog)
		if stopErr != nil {
			result.Result = *stopErr
			break
		}
		break
	}

	fn.PackageHTTPResponse(res, result)
}

func (s *Docker) QueryHandle(_ context.Context, res http.ResponseWriter, req *http.Request) {
	result := &common.QueryServiceResult{}
	for {
		param := &common.ServiceParam{}
		err := fn.ParseJSONBody(req, nil, param)
		if err != nil {
			result.ErrorCode = cd.IllegalParam
			result.Reason = "非法参数"
			break
		}
		serviceInfo, serviceErr := s.bizPtr.Query(param.Name, param.Catalog)
		if serviceErr != nil {
			result.Result = *serviceErr
			break
		}

		result.ServiceInfo = serviceInfo
		break
	}

	fn.PackageHTTPResponse(res, result)
}
