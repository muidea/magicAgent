package service

import (
	"context"
	"net/http"

	cd "github.com/muidea/magicCommon/def"
	fn "github.com/muidea/magicCommon/foundation/net"

	engine "github.com/muidea/magicEngine"

	"github.com/muidea/magicAgent/internal/core/module/mariadb/biz"
	"github.com/muidea/magicAgent/pkg/common"
)

// Mariadb BaseService
type Mariadb struct {
	routeRegistry engine.Router

	bizPtr *biz.Mariadb

	endpointName string
}

// New create base
func New(endpointName string, bizPtr *biz.Mariadb) *Mariadb {
	ptr := &Mariadb{
		endpointName: endpointName,
		bizPtr:       bizPtr,
	}

	return ptr
}

func (s *Mariadb) BindRegistry(
	routeRegistry engine.Router) {

	s.routeRegistry = routeRegistry

	s.routeRegistry.SetApiVersion(common.ApiVersion)
}

// RegisterRoute 注册路由
func (s *Mariadb) RegisterRoute() {
	statusRoute := engine.CreateRoute(common.QueryStatus, engine.GET, s.QueryStatusHandle)
	s.routeRegistry.AddRoute(statusRoute)
}

func (s *Mariadb) QueryStatusHandle(_ context.Context, res http.ResponseWriter, req *http.Request) {
	result := &common.QueryClusterStatusResult{}
	for {
		serviceName := req.URL.Query().Get("service")
		if serviceName == "" {
			result.ErrorCode = cd.IllegalParam
			result.Reason = "illegal service name"
			break
		}

		statusPtr, statusErr := s.bizPtr.QueryMariadbClusterStatus(serviceName)
		if statusErr != nil {
			result.Result = *statusErr
			break
		}

		result.Status = statusPtr
		break
	}

	fn.PackageHTTPResponse(res, result)
}
