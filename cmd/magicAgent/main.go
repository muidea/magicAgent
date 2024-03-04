package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/muidea/magicCommon/application"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicAgent/internal/config"
	"github.com/muidea/magicAgent/internal/core"
)

var listenPort = "8080"
var endpointName = "magicAgent"
var configFile = ""

func initPprofMonitor(listenPort string) {
	addr := ":1" + listenPort

	go func() {
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			log.Criticalf("funcRetErr=http.ListenAndServe||err=%s", err.Error())
		}
	}()
}

func main() {
	flag.StringVar(&listenPort, "ListenPort", listenPort, "listen address")
	flag.StringVar(&endpointName, "EndpointName", endpointName, "endpoint name.")
	flag.StringVar(&configFile, "Config", configFile, "config file path")
	flag.Parse()

	initPprofMonitor(listenPort)

	if configFile != "" {
		configErr := config.LoadConfig(configFile)
		if configErr != nil {
			log.Errorf("load config file failed, error:%s", configErr.Error())
			return
		}
	}

	fmt.Printf("%s starting!\n", endpointName)

	corePtr, coreErr := core.New(endpointName, listenPort)
	if coreErr != nil {
		log.Errorf("create core service failed, err:%s", coreErr.Error())
		return
	}
	err := application.Startup(corePtr)
	if err != nil {
		log.Errorf("application.Startup err:%s", err.Error())
		return
	}

	application.Run()
	application.Shutdown()
}
