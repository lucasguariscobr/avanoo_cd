package deploy

import (
	"avanoo_cd/utils"
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

var queue chan *Build

func StartBuildAgent() {
	queue = make(chan *Build)
	go buildQueue()
	unqueueBuilds()
}

func buildQueue() {
	for {
		select {
		case <-webhookCtx.Done():
			log.Printf("Ending Build Queue")
			return
		case currentBuild := <-queue:
			updateBuild(currentBuild.BuildId, utils.StatusStarted)
			buildErr := buildImage(currentBuild.Branch, &webhookCtx)
			currentBuild.err = buildErr
			currentBuild.wg.Done()
		}
	}
}

func buildImage(branchName string, ctx *context.Context) error {
	log.Printf("Start BuildImage")
	packer_command := buildCommand(branchName)
	cmd := exec.CommandContext(*ctx, "/bin/bash", "-c", packer_command)
	cmd.Dir = utils.PlaybookPath(branchName)
	outByte, err := utils.ExecCommand(utils.BuildCommand{Commandable: cmd})
	var output strings.Builder
	output.Write(outByte)

	if err != nil {
		sendBuildErrorEmail(branchName, cmd.String(), output.String(), err.Error())
	}
	return err
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
			domains := findBuildDomains(build)
			deploy(build, domains)
		}
	}
}
