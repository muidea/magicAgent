package core

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/module"
	"github.com/muidea/magicCommon/task"

	engine "github.com/muidea/magicEngine"

	"github.com/muidea/magicAgent/pkg/common"

	_ "github.com/muidea/magicAgent/internal/core/kernel/base"
	_ "github.com/muidea/magicAgent/internal/core/module/docker"
)

// New 新建Core
func New(endpointName, listenPort string) (ret *Core, err error) {
	core := &Core{
		endpointName: endpointName,
		listenPort:   listenPort,
	}

	ret = core
	return
}

// Core Core对象
type Core struct {
	endpointName string
	listenPort   string

	httpServer        engine.HTTPServer
	eventHub          event.Hub
	backgroundRoutine task.BackgroundRoutine
}

// Startup 启动
func (s *Core) Startup(eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) *cd.Result {
	router := engine.NewRouter()
	s.httpServer = engine.NewHTTPServer(s.listenPort)
	s.httpServer.Bind(router)

	modules := module.GetModules()
	for _, val := range modules {
		module.BindRegistry(val, router)
		module.Setup(val, s.endpointName, eventHub, backgroundRoutine)
	}

	s.eventHub = eventHub
	s.backgroundRoutine = backgroundRoutine
	return nil
}

func (s *Core) Run() {
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.httpServer.Run()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		modules := module.GetModules()
		for _, val := range modules {
			module.Run(val)
		}

		ev := event.NewEvent(common.NotifyRunning, "/", "/#", nil, nil)
		s.eventHub.Post(ev)
	}()

	wg.Wait()
}

// Shutdown 销毁
func (s *Core) Shutdown() {
	modules := module.GetModules()
	totalSize := len(modules)
	for idx := range modules {
		module.Teardown(modules[totalSize-idx-1])
	}
}
