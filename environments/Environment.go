package environments

import (
	"avanoo_cd/utils"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
)

type Environment interface {
	updateEnvironment(comm *utils.DomainComm, wg *sync.WaitGroup)
}

func UpdateEnvironment(buildComm *utils.BranchComm, domainName string, branchName string, hostName string, extraVars map[string]string) error {
	var environment Environment
	var environmentWG sync.WaitGroup
	comm := utils.DomainComm{}
	comm.ServicesStatus = &[]*utils.DeployStatus{}
	comm.Ctx = buildComm.Ctx
	comm.WG = &environmentWG
	comm.DomainName = domainName

	defer buildComm.WG.Done()
	defer completeEnvironmentDeploy(buildComm, &comm, branchName)

	log.Printf("%v - %v", domainName, branchName)
	switch domainName {
	case "app.avanoo.com":
	case "pre.avanoo.com":
		environment = createStageEnvironment(domainName, branchName, hostName, extraVars)
	default:
		environment = createBasicEnvironment(domainName, branchName, hostName, extraVars)
	}

	environment.updateEnvironment(&comm, &environmentWG)
	log.Printf("Endend environment update: %v", domainName)
	return nil
}

func updateService(comm *utils.DomainComm, serviceName string, command string, playbook string, branchName string, hostName string, extraVars map[string]string) error {
	serviceStatus := &utils.DeployStatus{
		ServiceName: serviceName,
		Status:      utils.StatusDeploySuccessful,
	}
	defer comm.WG.Done()

	log.Printf("Start %s service update", serviceName)
	command = fmt.Sprintf(command, hostName)
	updateDomainDeployStatus(comm, serviceStatus)
	builtCommand := buildCommand(command, playbook, branchName, extraVars)
	cmd := exec.CommandContext(*comm.Ctx, "/bin/bash", "-c", builtCommand)
	cmd.Dir = utils.PlaybookPath(branchName)
	outByte, err := utils.ExecCommand(utils.DeployCommand{Commandable: cmd})
	var output strings.Builder
	output.Write(outByte)
	serviceStatus.Command = cmd.String()
	serviceStatus.Output = output.String()

	if err != nil {
		log.Printf("command: %s", cmd.String())
		serviceStatus.Status = utils.StatusDeployError
		serviceStatus.ErrorMsg = err.Error()
	}
	log.Printf("%s service update finished with error: %v", serviceName, err)
	return err
}

func buildCommand(command string, playbook string, branchName string, extraVars map[string]string) string {
	var commandBuilder strings.Builder
	extraVarsList := buildExtraVarsList(extraVars, branchName)
	commandBuilder.WriteString("ANSIBLE_STDOUT_CALLBACK=debug ansible-playbook ")
	commandBuilder.WriteString(command)

	if len(extraVarsList) > 0 {
		commandBuilder.WriteString(" --extra-vars \"")
		commandBuilder.WriteString(strings.Join(extraVarsList, " "))
		commandBuilder.WriteString("\"")
	}
	commandBuilder.WriteString(" " + playbook)
	return commandBuilder.String()
}

func buildExtraVarsList(extraVars map[string]string, branchName string) []string {
	extraVarsList := []string{}
	for k, v := range extraVars {
		extraVarsList = append(extraVarsList, k+"="+v)
	}
	return extraVarsList
}

func completeEnvironmentDeploy(buildComm *utils.BranchComm, deployComm *utils.DomainComm, branchName string) {
	var resume string
	responseStatus := utils.CheckServicesStatusList(*deployComm.ServicesStatus)
	updateBranchDeployStatus(buildComm, &responseStatus)
	if responseStatus == utils.StatusDeploySuccessful {
		resume = "Domain  " + deployComm.DomainName + " was updated successfully with branch " + branchName
	} else {
		resume = "Error updating domain " + deployComm.DomainName + " with branch " + branchName
	}
	utils.SendEmail(&utils.EmailTemplateData{
		Subject:  "Avanoo CD:  " + resume,
		Resume:   resume,
		Services: *deployComm.ServicesStatus,
	})
}

func updateBranchDeployStatus(comm *utils.BranchComm, environmentStatus *string) {
	newBuildStatus := append(*comm.BuildStatus, *environmentStatus)
	comm.BuildStatus = &newBuildStatus
}

func updateDomainDeployStatus(comm *utils.DomainComm, serviceStatus *utils.DeployStatus) {
	newServicesStatus := append(*comm.ServicesStatus, serviceStatus)
	comm.ServicesStatus = &newServicesStatus
}
