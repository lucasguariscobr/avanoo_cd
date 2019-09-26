package main

import (
	"avanoo_cd/deploy"
	"avanoo_cd/server"
	"avanoo_cd/utils"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// @title Avanoo Continuous Delivery
// @version 1.0
// @description Automating cd process.

// @contact.name Lucas
// @contact.email lucas@avanoo.com

// @host control.placeboapp.com
// @BasePath /

func main() {
	closeFunc := utils.ReadConfig()
	closeWebHooksFunc := deploy.CreateDeployContext()
	appServer, errApp := server.CreateAppServer(utils.Address)
	if errApp != nil {
		log.Fatalf("Server Start Error: %v\n", errApp.Error())
	}
	appServer.InitServer()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

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
}
