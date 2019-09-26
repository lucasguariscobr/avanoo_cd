package environments

import (
	"avanoo_cd/utils"
	"sync"
)

type BasicEnvironment struct {
	DomainName string
	BranchName string
	HostName   string
	ExtraVars  map[string]string
}

func createBasicEnvironment(domainName string, branchName string, hostName string, extraVars map[string]string) Environment {
	return &BasicEnvironment{
		DomainName: domainName,
		BranchName: branchName,
		HostName:   hostName,
		ExtraVars:  extraVars,
	}
}

func (env *BasicEnvironment) updateEnvironment(comm *utils.DomainComm, environmentWG *sync.WaitGroup) {
	environmentWG.Add(2)
	go updateService(comm,
		"web",
		"--inventory=inventories/inventory.yml --limit %s --tags \"app_update\"",
		" app_provision.yml",
		env.BranchName,
		env.HostName,
		env.ExtraVars)
	go updateService(comm,
		"background",
		"--inventory=inventories/inventory.yml --limit %s --tags \"sidekiq_update\"",
		" app_provision.yml",
		env.BranchName,
		env.HostName,
		env.ExtraVars)
	environmentWG.Wait()
}
