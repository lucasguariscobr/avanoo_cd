package deploy

import (
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
var buildMap map[string]*Build

func CreateDeployContext() func() {
	buildMap = make(map[string]*Build)
	webhookCtx, webhookCancelFunc = context.WithCancel(context.Background())
	return func() {
		log.Printf("Closing Deploy Context")
		webhookCancelFunc()
		deployWG.Wait()
		buildImageWG.Wait()
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

	releaseBranch(branchName, buildDomains)
}

func releaseBranch(branchName string, buildDomains []*Domain) {
	uuid, _ := uuid2.NewV4()
	build, err := createBuild(uuid.String(), branchName, buildDomains)
	if err != nil {
		log.Printf(err.Error())
		return
	}
	cancelWaitingBuild(branchName)
	cancelRunningBuildImage(branchName)
	buildMap[build.BuildId] = build
	deploy(build, buildDomains)
}

func deploy(build *Build, domains []*Domain) {
	var buildWG sync.WaitGroup
	build.wg = &buildWG
	build.context, build.ctxCancelFunc = context.WithCancel(webhookCtx)
	buildWG.Add(1)
	buildQueue <- build

	buildWG.Wait()
	if build.err != nil {
		delete(buildMap, build.BuildId)
		webhookWG.Done()
		return
	}

	deployQueue <- build
}

func cancelRunningBuildImage(branchName string) {
	for _, build := range buildMap {
		if build.Branch == branchName && build.Status == utils.StatusStarted {
			build.ctxCanceled = true
			build.ctxCancelFunc()
			break
		}
	}
}

func cancelWaitingBuild(branchName string) {
	for _, build := range buildMap {
		if build.Branch == branchName && (build.Status == utils.StatusQueued || build.Status == utils.StatusBuildCompleted) {
			updateBuild(build, utils.StatusCanceled)
			break
		}
	}
}