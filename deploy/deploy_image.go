package deploy

import (
	"context"
	"log"
	"sync"
	"avanoo_cd/utils"
	"avanoo_cd/environments"
)

var waitingMap map[string]*Build
var runningMap map[string]*Build
var deployQueue chan *Build
var deployWG sync.WaitGroup

func StartDeployAgent() {
	waitingMap = make(map[string]*Build)
	runningMap = make(map[string]*Build)
	deployQueue = make(chan *Build)
	for i:=0; i < 2; i ++ {
		go startDeployQueue()
	}
}

func startDeployQueue() {
	deployWG.Add(1)
	deployCtx, _ := context.WithCancel(webhookCtx)

	for {
		select {
		case <-deployCtx.Done():
			log.Printf("Ending Deploy Queue")
			deployWG.Done()
			return
		case build := <- deployQueue:
			log.Printf("Start Deploy")
			if build.Status == utils.StatusCanceled {
				continue
			}

			_, ok := runningMap[build.Branch]
			if !ok {
				runningMap[build.Branch] = build
				updateBuild(build, utils.StatusDeployStarted)
				runDeploy(build, &deployCtx)
			} else {
				waitingMap[build.Branch] = build
			}
		}
	}
}

func runDeploy(currentBuild *Build, deployCtx *context.Context) {
	defer webhookWG.Done()
	defer postDeploy(currentBuild.Branch)
	defer delete(runningMap, currentBuild.Branch)
	defer delete(buildMap, currentBuild.BuildId)
	
	var domainsWG sync.WaitGroup
	comm := utils.BranchComm{}
	comm.BuildStatus = &[]string{}
	comm.Ctx = deployCtx
	comm.WG = &domainsWG
	
	defer completeBuild(currentBuild, &comm)	
	
	domainsWG.Add(len(currentBuild.Domains))
	for _, domain := range currentBuild.Domains {
		go environments.UpdateEnvironment(&comm, domain.Domain, domain.Branch, domain.Host, domain.ExtraVars)
	}
	domainsWG.Wait()
	log.Printf("Stop environments update")
}

func postDeploy(branch string) {
	build, ok := waitingMap[branch]
	if ok {
		delete(waitingMap, branch)
		deployQueue <- build
	}
}