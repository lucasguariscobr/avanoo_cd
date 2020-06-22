package deploy

import (
	"avanoo_cd/utils"
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
)

var buildQueue chan *Build
var buildCtx context.Context
var buildImageWG sync.WaitGroup

func StartBuildAgent() {
	buildQueue = make(chan *Build)
	buildCtx, _ = context.WithCancel(webhookCtx)
	go startBuildQueue()
	unqueueBuilds()
}

func startBuildQueue() {
	buildImageWG.Add(1)
	for {
		select {
		case <-buildCtx.Done():
			log.Printf("Ending Build Queue")
			buildImageWG.Done()
			return
		case currentBuild := <-buildQueue:
			updateBuild(currentBuild, utils.StatusStarted)
			buildImage(currentBuild)
			currentBuild.wg.Done()
		}
	}
}

func postBuildImage(currentBuild *Build, err error) {
	if currentBuild.ctxCanceled == true {
		updateBuild(currentBuild, utils.StatusCanceled)
	} else if err == nil {
		updateBuild(currentBuild, utils.StatusBuildCompleted)
	} else {
		log.Printf(err.Error())
		updateBuild(currentBuild, utils.StatusBuildError)
	}
	currentBuild.err = err
}

func buildImage(currentBuild *Build) {
	log.Printf("Start BuildImage")
	packer_command := buildCommand(currentBuild.Branch)
	cmd := exec.CommandContext(currentBuild.context, "/bin/bash", "-c", packer_command)
	cmd.Dir = utils.PlaybookPath(currentBuild.Branch)
	outByte, err := utils.ExecCommand(utils.BuildCommand{Commandable: cmd})
	var output strings.Builder
	output.Write(outByte)

	if err != nil && currentBuild.ctxCanceled == false {
		sendBuildErrorEmail(currentBuild.Branch, cmd.String(), output.String(), err.Error())
	}

	postBuildImage(currentBuild, err)
}

func buildCommand(branchName string) string {
	var commandBuilder strings.Builder
	commandBuilder.WriteString(fmt.Sprintf("packer build -var \"BRANCH=%s\" app_docker.json ", branchName))
	return commandBuilder.String()
}

func sendBuildErrorEmail(branchName string, command string, output string, errorMsg string) {
	buildStatus := &utils.DeployStatus{
		ServiceName: "build",
		Command:     command,
		Output:      output,
		ErrorMsg:    errorMsg,
	}

	resume := "Error building image " + branchName
	utils.SendEmail(&utils.EmailTemplateData{
		Subject:  "Avanoo CD:  " + resume,
		Resume:   resume,
		Services: []*utils.DeployStatus{buildStatus},
	})
}

func unqueueBuilds() {
	builds, err := scanBuilds()
	if err != nil {
		log.Printf(err.Error())
	}

	for _, build := range builds {
		if build.Status == utils.StatusQueued {
			webhookWG.Add(1)
			build.Domains = findBuildDomains(build)
			deploy(build, build.Domains)
		}
	}
}
