package utils

import (
	"context"
	"sync"
)

type DeployStatus struct {
	ServiceName string `json:"service_name"`
	Status      string `json:"-"`
	Command     string `json:"command"`
	Output      string `json:"output"`
	Error       bool   `json:"error"`
	ErrorMsg    string `json:"error_msg"`
}

type DomainComm struct {
	ServicesStatus *[]*DeployStatus
	Ctx            *context.Context
	WG             *sync.WaitGroup
	DomainName     string
}

type BranchComm struct {
	BuildStatus *[]string
	Ctx         *context.Context
	WG          *sync.WaitGroup
}

const (
	StatusQueued           = "Queued"
	StatusStarted          = "Started"
	StatusBuildError       = "Build Error"
	StatusBuildCompleted   = "Build Complete"
	StatusDeployStarted	   = "Deploy Started"
	StatusDeploySuccessful = "Deploy Successful"
	StatusDeployError      = "Deploy Error"
	StatusCanceled		   = "Canceled"
)

func CheckServicesStatusList(deployStatusList []*DeployStatus) string {
	statusList := []string{}
	for _, status := range deployStatusList {
		statusList = append(statusList, status.Status)
	}
	return CheckStatusList(statusList)
}

func CheckStatusList(statusList []string) string {
	responseStatus := StatusDeploySuccessful
	if len(statusList) == 0 {
		responseStatus = StatusDeployError
	}
	for _, status := range statusList {
		if status == StatusDeployError {
			responseStatus = StatusDeployError
			break
		}
	}

	return responseStatus
}
