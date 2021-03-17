package environments

import (
	"github.com/avanoo/avanoo_cd/utils"
	"sync"
)

type StageEnvironment struct {
	DomainName string
	BranchName string
	HostName   string
	ExtraVars  map[string]string
}

func createStageEnvironment(domainName string, brachName string, hostName string, extraVars map[string]string) Environment {
	return &StageEnvironment{
		DomainName: domainName,
		BranchName: brachName,
		HostName:   hostName,
		ExtraVars:  extraVars,
	}
}

func (env *StageEnvironment) updateEnvironment(comm *utils.DomainComm, environmentWG *sync.WaitGroup) {
	environmentWG.Add(2)

	webExtraVars := map[string]string{}
	backgroundbExtraVars := map[string]string{}
	for k, v := range env.ExtraVars {
		webExtraVars[k] = v
		backgroundbExtraVars[k] = v
	}

	webExtraVars["AUTO_SCALING_TARGET_GROUP"] = "stage"
	webExtraVars["AUTOSCALING_BRANCH"] = env.BranchName
	go updateService(comm,
		"web",
		"--inventory=inventories/ec2.py -l %s -u ubuntu --private-key \"/home/avanoo/.ssh/id_rsa\" --tags \"app_update\"",
		" deploy.yml",
		env.BranchName,
		"tag_aws_autoscaling_groupName_stage",
		webExtraVars)

	backgroundbExtraVars["AUTO_SCALING_TARGET_GROUP"] = "stage-bg"
	backgroundbExtraVars["AUTOSCALING_BRANCH"] = env.BranchName
	go updateService(comm,
		"background",
		"--inventory=inventories/ec2.py -l %s -u ubuntu --private-key \"/home/avanoo/.ssh/id_rsa\" --tags \"sidekiq_update\"",
		" deploy.yml",
		env.BranchName,
		"tag_aws_autoscaling_groupName_stage_bg",
		backgroundbExtraVars)
	environmentWG.Wait()
}
