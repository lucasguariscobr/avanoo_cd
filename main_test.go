package main

import (
	"avanoo_cd/utils"
	"net/http"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	utils.SetEnv("test")
	utils.SetConfigurationFlag()
	result := m.Run()
	os.Exit(result)
}

func TestUnknownEndpoint(t *testing.T) {
	startMainServer()
	response, _ := http.Get("http://" + utils.Address + "/unknown")
	status := response.StatusCode
	if status != 404 {
		t.Error("Unknown endpoint error")
	}
	stopMainServer()
}

func startMainServer() {
	testWG.Add(1)
	go main()
	testWG.Wait()
}

func stopMainServer() {
	testWG.Add(1)
	pid := os.Getpid()
	p, _ := os.FindProcess(pid)
	p.Signal(os.Interrupt)
	testWG.Wait()
}
