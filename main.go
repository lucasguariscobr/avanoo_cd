package main

import (
	"github.com/avanoo/avanoo_cd/deploy"
	"github.com/avanoo/avanoo_cd/server"
	"github.com/avanoo/avanoo_cd/utils"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var testWG sync.WaitGroup

// @title Avanoo Continuous Delivery
// @version 1.0
// @license.name GNU GPLv3
// @description Automating cd process.
// @contact.name Lucas
// @contact.email lucas@avanoo.com
// @host cd.placeboapp.com
func main() {
	closeFunc := utils.ReadConfig()
	closeWebHooksFunc := deploy.CreateDeployContext()
	deploy.StartDeployAgent()
	deploy.StartBuildAgent()
	appServer, errApp := server.CreateAppServer(utils.Address)
	if errApp != nil {
		log.Fatalf("Server Start Error: %v\n", errApp.Error())
	}
	appServer.InitServer()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	checkpoint()

	select {
	case <-stop:
		log.Printf("Stop Signal \n")
	case <-appServer.ErrorChannel:
		log.Printf("Error Signal App Server \n")
	}

	appServer.CloseApplication()
	closeWebHooksFunc()
	closeFunc()
	log.Printf("Final Shutdown\n")
	checkpoint()
}

func checkpoint() {
	if utils.Env == "test" {
		testWG.Done()
	}
}
