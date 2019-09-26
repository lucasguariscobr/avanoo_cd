package deploy

import (
	"avanoo_cd/utils"
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func buildImage(branchName string, ctx *context.Context) error {
	log.Printf("Start BuildImage")
	packer_command := buildCommand(branchName)
	cmd := exec.CommandContext(*ctx, "/bin/bash", "-c", packer_command)
	outByte, err := cmd.Output()
	var output strings.Builder
	output.Write(outByte)

	if err != nil {
		sendBuildErrorEmail(branchName, cmd.String(), output.String(), err.Error())
	}
	return err
}

func buildCommand(branchName string) string {
	var commandBuilder strings.Builder
	playbook_path := utils.PlaybookPath(branchName)
	commandBuilder.WriteString(fmt.Sprintf("cd %s && ", playbook_path))
	commandBuilder.WriteString(fmt.Sprintf("packer build -var \"ENV=%s\" app_docker.json ", branchName))
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
