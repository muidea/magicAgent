package service

import (
	"github.com/muidea/magicAgent/internal/core/kernel/base/biz"
	"github.com/muidea/magicAgent/pkg/common"
	engine "github.com/muidea/magicEngine"
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

// RegisterRoute 注册路由
func (s *Base) RegisterRoute() {
}
