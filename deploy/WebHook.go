package deploy

import (
	"avanoo_cd/environments"
	"avanoo_cd/utils"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	uuid2 "github.com/satori/go.uuid"
)

type PushJSON struct {
	Ref string `json:"ref"`
}

type Build struct {
	BuildId string
	Branch  string
	Domains []string
	Date    string
	Status  string
}

var buildNamespace = "build_"
var githubRef, _ = regexp.Compile("/ref/heads/(.*)")
var buildExpiration, _ = time.ParseDuration("168h")
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

func ListenWebHookEvent(w http.ResponseWriter, r *http.Request) {
	webhookWG.Add(1)
	pushJSON, err := utils.DecodeMsg(r, &PushJSON{}, false)
	if err != nil {
		utils.WriteJSONError(w, err.Error())
		webhookWG.Done()
		return
	}

	branch := githubRef.FindStringSubmatch(pushJSON.(*PushJSON).Ref)
	branchName := branch[1]
	go verifyBranch(branchName)
	w.WriteHeader(http.StatusNoContent)
}

func ListBuilds(w http.ResponseWriter, r *http.Request) {
	builds, err := scanBuilds()
	if err != nil {
		log.Printf(err.Error())
		utils.WriteJSONError(w, "It wasn't possible to fetch the list of builds")
		return
	}

	buildValues := []Build{}
	for _, value := range builds {
		buildValues = append(buildValues, *value)
	}

	json.NewEncoder(w).Encode(buildValues)
}

func verifyBranch(branchName string) error {
	var buildDomains []*DomainJSON
	var domainsWG sync.WaitGroup
	uuid, _ := uuid2.NewV4()
	defer webhookWG.Done()

	domains, err := scanDomains()
	if err != nil {
		log.Printf(err.Error())
		return err
	}

	for _, value := range domains {
		if value.Branch == branchName {
			buildDomains = append(buildDomains, value)
		}
	}

	if len(buildDomains) == 0 {
		return nil
	}

	comm := utils.BranchComm{}
	comm.BuildStatus = &[]string{}
	comm.Ctx = &webhookCtx
	comm.WG = &domainsWG
	recordBuild(uuid.String(), branchName, buildDomains)

	errBuild := buildImage(branchName, &webhookCtx)
	if errBuild == nil {
		defer completeBuild(uuid.String(), &comm)

		log.Printf("Start environments update")
		updateBuild(uuid.String(), utils.StatusBuildCompleted)
		domainsWG.Add(len(buildDomains))
		for _, domain := range buildDomains {
			go environments.UpdateEnvironment(&comm, domain.Domain, domain.Branch, domain.Host, domain.ExtraVars)
		}
		domainsWG.Wait()
		log.Printf("Stop environments update")
	} else {
		log.Printf(errBuild.Error())
		updateBuild(uuid.String(), utils.StatusBuildError)
		return errBuild
	}

	return nil
}

func scanBuilds() (map[string]*Build, error) {
	builds := map[string]*Build{}
	keyIter := utils.RedisClient.Scan(0, buildNamespace+"*", 100).Iterator()
	for keyIter.Next() {
		buildId := keyIter.Val()
		build, err := fetchBuildInfo(buildId)
		if err != nil {
			return nil, err
		}
		builds[buildId] = build
	}
	if err := keyIter.Err(); err != nil {
		return nil, err
	}

	return builds, nil
}

func fetchBuildInfo(buildId string) (*Build, error) {
	buildInfo, errRedis := utils.RedisClient.HMGet(buildId, "id", "branch", "date", "domains", "status").Result()
	if errRedis != nil {
		return nil, errRedis
	}
	build := Build{}
	if len(buildInfo) < 3 {
		return nil, errors.New("no build info")
	}
	log.Printf("Build Info: %v", buildInfo)
	build.BuildId = buildInfo[0].(string)
	build.Branch = buildInfo[1].(string)
	build.Date = buildInfo[2].(string)
	domain_string := buildInfo[3].(string)
	build.Domains = strings.Split(domain_string, ",")
	build.Status = buildInfo[4].(string)

	return &build, nil
}

func recordBuild(id string, branchName string, domains []*DomainJSON) error {
	time := time.Now().Format("2006-01-02 15:04:05")
	domainNames := []string{}
	for _, domain := range domains {
		domainNames = append(domainNames, domain.Domain)
	}

	newBuild := map[string]interface{}{
		"id":      id,
		"branch":  branchName,
		"date":    time,
		"domains": strings.Join(domainNames, ","),
		"status":  utils.StatusStarted,
	}
	buildKey := buildNamespace + id
	err := utils.RedisClient.HMSet(buildKey, newBuild).Err()
	if err != nil {
		return err
	}

	err = utils.RedisClient.Expire(buildKey, buildExpiration).Err()
	if err != nil {
		return err
	}
	return err
}

func updateBuild(id string, status string) error {
	log.Printf("Update Build: %s", status)
	buildKey := buildNamespace + id
	buildInfo, errFetch := fetchBuildInfo(buildKey)
	if errFetch != nil {
		return errFetch
	}

	time := time.Now().Format("2006-01-02 15:04:05")
	updateBuild := map[string]interface{}{
		"id":      buildKey,
		"branch":  buildInfo.Branch,
		"date":    time,
		"domains": strings.Join(buildInfo.Domains, ","),
		"status":  status,
	}
	err := utils.RedisClient.HMSet(buildKey, updateBuild).Err()
	return err
}

func completeBuild(id string, comm *utils.BranchComm) error {
	responseStatus := utils.CheckStatusList(*comm.BuildStatus)
	log.Printf("Build is complete %v", responseStatus)
	return updateBuild(id, responseStatus)
}
