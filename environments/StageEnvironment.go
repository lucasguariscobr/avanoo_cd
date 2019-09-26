package environments

import (
	"avanoo_cd/utils"
	"sync"
)

type StageEnvironment struct {
	DomainName string
	BranchName string
	HostName   string
	ExtraVars  map[string]string
}

func createStageEnvironment(domainName string, brachName string, hostName string, extraVars map[string]string) Environment {
	return &BasicEnvironment{
		DomainName: domainName,
		BranchName: brachName,
		HostName:   hostName,
		ExtraVars:  extraVars,
	}
}

func (env *StageEnvironment) updateEnvironment(comm *utils.DomainComm, environmentWG *sync.WaitGroup) {
	environmentWG.Add(2)
	go updateService(comm,
		"web",
		"--inventory=inventories/ec2.py -l %s -u ubuntu --private-key \"/home/avanoo/.ssh/id_rsa\" --tags \"app_update\"",
		" app_provision.yml",
		env.BranchName,
		env.HostName,
		env.ExtraVars)
	go updateService(comm,
		"background",
		"--inventory=inventories/ec2.py -l %s -u ubuntu --private-key \"/home/avanoo/.ssh/id_rsa\" --tags \"sidekiq_update\"",
		" app_provision.yml",
		env.BranchName,
		env.HostName,
		env.ExtraVars)
	environmentWG.Wait()
}
