package deploy

import (
	"github.com/avanoo/avanoo_cd/utils"
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
	for i:=0; i < utils.BuildImageAgents; i ++ {
		go startBuildQueue()
	}
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
			runBuild(currentBuild)
			currentBuild.wg.Done()
		}
	}
}

func runBuild(currentBuild *Build) {
	var runBuildWG sync.WaitGroup

	packerExec, packerArgs := buildDockerCommand(currentBuild.Branch)
	err, ok := execBuildCommand(currentBuild, packerExec, packerArgs)
	if !ok {
		postBuildImage(currentBuild, err)
		return
	}

	environments := getDomainEnvironments(currentBuild)
	for host, environment := range environments {
		runBuildWG.Add(1)
		go pre_deploy(currentBuild, &runBuildWG, environment, host)
	}
	runBuildWG.Wait()
}


func buildDockerCommand(branchName string) (string, []string) {
	packerExecutable := "/bin/bash"
	commandBuilder := []string{"-c"}
	commandBuilder = append(commandBuilder, fmt.Sprintf("packer build -var \"BRANCH=%s\" app_docker.json", branchName))
	return packerExecutable, commandBuilder
}

func pre_deploy(currentBuild *Build, wg *sync.WaitGroup, environment string, domain string ) {
	defer wg.Done()

	ansibleExec, ansibleArgs := buildPreDeployCommand(currentBuild.Branch, environment, domain)
	err, _ := execBuildCommand(currentBuild, ansibleExec, ansibleArgs)
	if currentBuild.Status == utils.StatusCanceled || currentBuild.Status == utils.StatusBuildError {
		return
	}
	postBuildImage(currentBuild, err)
}

func buildPreDeployCommand(branchName string, environment string, domain string) (string, []string) {
	ansibleExecutable := "/bin/bash"
	commandBuilder := []string{"-c"}
	commandBuilder = append(commandBuilder, fmt.Sprintf("ansible-playbook --inventory=127.0.0.1, -c local -e \"BRANCH=%s RAILS_ENV=%s DOMAIN_NAME=%s\" pre_deploy.yml", branchName, environment, domain))
	return ansibleExecutable, commandBuilder
}

func execBuildCommand(currentBuild *Build, commandExecutable string, commandArgs []string) (error, bool) {
	cmd := exec.CommandContext(currentBuild.context, commandExecutable, commandArgs...)
	cmd.Dir = utils.PlaybookPath(currentBuild.Branch)
	outByte, err := utils.ExecCommand(utils.BuildCommand{Commandable: cmd})
	var output strings.Builder
	output.Write(outByte)

	if err != nil && currentBuild.ctxCanceled == false {
		sendBuildErrorEmail(currentBuild.Branch, cmd.String(), output.String(), err.Error())
		return err, false
	}
	
	return err, true
}

func getDomainEnvironments(currentBuild *Build) map[string]string {
	environments := []string{}
	environmentMap := map[string]string{}
	var ok bool
	for _, domain := range currentBuild.Domains {
		_, ok = environmentMap[domain.Host]
		if !ok {
			switch(domain.Host) {
			case "pre": 
				environmentMap[domain.Host] = "stage"	
			case "prodtest": 
				environmentMap[domain.Host] = "production"
			}
		}
	}

	return environmentMap
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
