package deploy

import (
	"avanoo_cd/environments"
	"avanoo_cd/utils"
	"context"
	"log"
	"net/http"
	"regexp"
	"sync"

	uuid2 "github.com/satori/go.uuid"
)

type PushJSON struct {
	Ref     string `json:"ref"`
	Deleted bool   `json:"deleted"`
}

var githubRef, _ = regexp.Compile("refs/heads/(.*)")
var webhookCtx context.Context
var webhookCancelFunc func()
var webhookWG sync.WaitGroup

func CreateDeployContext() func() {
	webhookCtx, webhookCancelFunc = context.WithCancel(context.Background())
	return func() {
		log.Printf("Closing Deploy Context")
		webhookCancelFunc()
		webhookWG.Wait()
		log.Printf("Closed Deploy Context")
	}
}

// ListenWebHookEvent godoc
// @Summary GitHub integration
// @Description Webhook that communicates with Github.
// @Description Currently listens to push events from the App repository.
// @Tags webhook
// @Success 204
// @Failure 400 {object} utils.JSONErrror
// @Failure 404
// @Failure 405
// @Router /webhook [post]
func ListenWebHookEvent(w http.ResponseWriter, r *http.Request) {
	webhookWG.Add(1)
	pushJSON, err := utils.DecodeMsg(r, &PushJSON{}, false)
	if err != nil {
		utils.WriteJSONError(w, err.Error())
		webhookWG.Done()
		return
	}

	push := pushJSON.(*PushJSON)
	if push.Ref == "" {
		utils.WriteJSONError(w, "Wrong webhook format")
		webhookWG.Done()
		return
	}

	if push.Deleted {
		w.WriteHeader(http.StatusNoContent)
		webhookWG.Done()
		return
	}

	branch := githubRef.FindStringSubmatch(push.Ref)
	if len(branch) == 0 {
		utils.WriteJSONError(w, "Invalid reference")
		webhookWG.Done()
		return
	}

	branchName := branch[1]
	go verifyBranch(branchName)
	w.WriteHeader(http.StatusNoContent)
}

func verifyBranch(branchName string) {
	var buildDomains []*Domain

	domains, err := scanDomains()
	if err != nil {
		log.Printf(err.Error())
		webhookWG.Done()
		return
	}

	for _, value := range domains {
		if value.Branch == branchName {
			buildDomains = append(buildDomains, value)
		}
	}

	if len(buildDomains) == 0 {
		webhookWG.Done()
		return
	}

	if duplicatedDeploy(branchName) {
		webhookWG.Done()
		return
	}

	releaseBranch(branchName, buildDomains)
}

func releaseBranch(branchName string, buildDomains []*Domain) {
	uuid, _ := uuid2.NewV4()
	build, err := createBuild(uuid.String(), branchName, buildDomains)
	if err != nil {
		log.Printf(err.Error())
		return
	}
	deploy(build, buildDomains)
}

func deploy(build *Build, domains []*Domain) {
	defer webhookWG.Done()

	var domainsWG sync.WaitGroup
	var buildWG sync.WaitGroup

	comm := utils.BranchComm{}
	comm.BuildStatus = &[]string{}
	comm.Ctx = &webhookCtx
	comm.WG = &domainsWG

	build.wg = &buildWG
	buildWG.Add(1)
	queue <- build

	buildWG.Wait()
	errBuild := build.err
	if errBuild == nil {
		defer completeBuild(build.BuildId, &comm)

		log.Printf("Start environments update")
		updateBuild(build.BuildId, utils.StatusBuildCompleted)
		domainsWG.Add(len(domains))
		for _, domain := range domains {
			go environments.UpdateEnvironment(&comm, domain.Domain, domain.Branch, domain.Host, domain.ExtraVars)
		}
		domainsWG.Wait()
		log.Printf("Stop environments update")
	} else {
		log.Printf(errBuild.Error())
		updateBuild(build.BuildId, utils.StatusBuildError)
	}
}

func duplicatedDeploy(branchName string) bool {
	builds, err := scanBuilds()
	if err != nil {
		log.Printf(err.Error())
		return false
	}

	for _, build := range builds {
		if build.Status == utils.StatusQueued && build.Branch == branchName {
			return true
		}
	}
	return false
}
