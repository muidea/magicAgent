package service

import (
	"context"
	"net/http"

	engine "github.com/muidea/magicEngine"

	cd "github.com/muidea/magicCommon/def"
	fn "github.com/muidea/magicCommon/foundation/net"

	"github.com/muidea/magieAgent/internal/config"
	"github.com/muidea/magieAgent/internal/core/kernel/base/biz"
	"github.com/muidea/magieAgent/pkg/common"
)

// Base BaseService
type Base struct {
	routeRegistry engine.Router

	bizPtr *biz.Base

	endpointName string
}

// New create base
func New(endpointName string, bizPtr *biz.Base) *Base {
	ptr := &Base{
		endpointName: endpointName,
		bizPtr:       bizPtr,
	}

	return ptr
}

func (s *Base) BindRegistry(
	routeRegistry engine.Router) {

	s.routeRegistry = routeRegistry

	s.routeRegistry.SetApiVersion(common.ApiVersion)
}

func (s *Base) Handle(ctx engine.RequestContext, res http.ResponseWriter, req *http.Request) {
	apiKey := req.Header.Get(common.APIKey)
	if apiKey != config.GetSyncSecret() {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	ctx.Next()
}

// RegisterRoute 注册路由
func (s *Base) RegisterRoute() {
	healthRoute := engine.CreateRoute(common.CheckHealth, engine.GET, s.Health)
	s.routeRegistry.AddRoute(healthRoute)

	haHealthRoute := engine.CreateRoute(common.CheckHAHealth, engine.GET, s.HAHealth)
	s.routeRegistry.AddRoute(haHealthRoute, s)

	haReadyRoute := engine.CreateRoute(common.CheckHAReady, engine.GET, s.HAReady)
	s.routeRegistry.AddRoute(haReadyRoute, s)

	syncRemoteRoute := engine.CreateRoute(common.SyncServiceDataToRemote, engine.POST, s.SyncToRemoteHandle)
	s.routeRegistry.AddRoute(syncRemoteRoute, s)

	syncLocalRoute := engine.CreateRoute(common.SyncServiceDataToLocal, engine.POST, s.SyncToLocalHandle)
	s.routeRegistry.AddRoute(syncLocalRoute, s)

	backupRoute := engine.CreateRoute(common.BackupMariadbData, engine.GET, s.BackupMariadbHandle)
	s.routeRegistry.AddRoute(backupRoute, s)

	restoreRoute := engine.CreateRoute(common.RestoreMariadbData, engine.GET, s.RestoreMariadbHandle)
	s.routeRegistry.AddRoute(restoreRoute, s)

	queryStatusRoute := engine.CreateRoute(common.QueryMariadbStatus, engine.GET, s.QueryMariadbStatusHandle)
	s.routeRegistry.AddRoute(queryStatusRoute, s)
}

func (s *Base) Health(_ context.Context, res http.ResponseWriter, _ *http.Request) {
	res.WriteHeader(http.StatusOK)
}

func (s *Base) HAHealth(_ context.Context, res http.ResponseWriter, _ *http.Request) {
	if s.bizPtr.GetKernelHealthStatus() {
		res.WriteHeader(http.StatusOK)
		return
	}

	res.WriteHeader(http.StatusForbidden)
}

func (s *Base) HAReady(_ context.Context, res http.ResponseWriter, _ *http.Request) {
	if s.bizPtr.GetKernelReadyStatus() {
		res.WriteHeader(http.StatusOK)
		return
	}

	res.WriteHeader(http.StatusForbidden)
}

func (s *Base) SyncToRemoteHandle(_ context.Context, res http.ResponseWriter, req *http.Request) {
	result := &common.SyncServiceDataResult{}
	for {
		_, serviceName := fn.SplitRESTPath(req.URL.Path)
		syncParam := &common.SyncInfo{}
		err := fn.ParseJSONBody(req, nil, syncParam)
		if err != nil {
			result.ErrorCode = cd.IllegalParam
			result.Reason = "非法参数"
			break
		}
		catalog := req.URL.Query().Get("catalog")
		syncErr := s.bizPtr.SyncDataToRemote(serviceName, catalog, syncParam)
		if syncErr != nil {
			result.Result = *syncErr
			break
		}

		break
	}

	fn.PackageHTTPResponse(res, result)
}

func (s *Base) SyncToLocalHandle(_ context.Context, res http.ResponseWriter, req *http.Request) {
	result := &common.SyncServiceDataResult{}
	for {
		_, serviceName := fn.SplitRESTPath(req.URL.Path)
		syncParam := &common.SyncInfo{}
		err := fn.ParseJSONBody(req, nil, syncParam)
		if err != nil {
			result.ErrorCode = cd.IllegalParam
			result.Reason = "非法参数"
			break
		}
		catalog := req.URL.Query().Get("catalog")
		syncErr := s.bizPtr.SyncDataToLocal(serviceName, catalog, syncParam)
		if syncErr != nil {
			result.Result = *syncErr
			break
		}

		break
	}

	fn.PackageHTTPResponse(res, result)
}

func (s *Base) BackupMariadbHandle(_ context.Context, res http.ResponseWriter, req *http.Request) {
	result := &common.BackupMariadbDataResult{}
	for {
		_, serviceName := fn.SplitRESTPath(req.URL.Path)
		dataPath, backupErr := s.bizPtr.BackupMariadb(serviceName)
		if backupErr != nil {
			result.Result = *backupErr
			break
		}

		result.DataPath = dataPath
		break
	}

	fn.PackageHTTPResponse(res, result)
}

func (s *Base) RestoreMariadbHandle(_ context.Context, res http.ResponseWriter, req *http.Request) {
	result := &common.RestoreMariadbDataResult{}
	for {
		_, serviceName := fn.SplitRESTPath(req.URL.Path)
		restoreErr := s.bizPtr.RestoreMariadb(serviceName)
		if restoreErr != nil {
			result.Result = *restoreErr
			break
		}
		break
	}

	fn.PackageHTTPResponse(res, result)
}

func (s *Base) QueryMariadbStatusHandle(_ context.Context, res http.ResponseWriter, req *http.Request) {
	result := &common.QueryMariadbStatusResult{}
	for {
		_, serviceName := fn.SplitRESTPath(req.URL.Path)
		masterStatus, slaveStatus, statusErr := s.bizPtr.QueryLocalMariadbStatus(serviceName)
		if statusErr != nil {
			result.Result = *statusErr
			break
		}

		result.MasterStatus = masterStatus
		result.SlaveStatus = slaveStatus
		break
	}

	fn.PackageHTTPResponse(res, result)
}
