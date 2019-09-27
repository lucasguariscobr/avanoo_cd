package deploy

import (
	"avanoo_cd/utils"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type Build struct {
	BuildId string   `json:"id"`
	Branch  string   `json:"branch"`
	Domains []string `json:"domains"`
	Date    string   `json:"date"`
	Status  string   `json:"status"`
	wg      *sync.WaitGroup
	err     error
}

const (
	FilterQueued     = "queued"
	FilterRunning    = "running"
	FilterCompleted  = "completed"
	FilterSuccessful = "successful"
	FilterFailed     = "failed"
)

var buildNamespace = "build_"
var buildExpiration, _ = time.ParseDuration("168h")

// ListBuilds godoc
// @Summary List all builds
// @Description Returns the list of builds executed by the service.
// @Description Only branches that are connected to a domain have their docker images generated.
// @Description The build information is kept for one week.
// @Param filter query string false "Filter Builds" Enums(queued, running, completed, successful, failed)
// @Tags builds
// @Produce  json
// @Success 200 {object} deploy.Build
// @Failure 400 {object} utils.JSONErrror
// @Failure 404
// @Failure 405
// @Router /builds [get]
func ListBuilds(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	statusFilterList := buildStatusFilter(filter)

	builds, err := scanBuilds()
	if err != nil {
		log.Printf(err.Error())
		utils.WriteJSONError(w, "It wasn't possible to fetch the list of builds")
		return
	}

	buildValues := []Build{}
	for _, value := range builds {
		if matchBuildStatus(statusFilterList, value) {
			buildValues = append(buildValues, *value)
		}
	}

	sort.SliceStable(buildValues,
		func(i, j int) bool {
			if buildValues[i].Date < buildValues[j].Date {
				return true
			}
			if buildValues[i].Date > buildValues[j].Date {
				return false
			}
			return buildValues[i].BuildId < buildValues[j].BuildId
		},
	)

	json.NewEncoder(w).Encode(buildValues)
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
		log.Printf("Build: %v", *build)
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
	build.BuildId = buildInfo[0].(string)
	build.Branch = buildInfo[1].(string)
	build.Date = buildInfo[2].(string)
	domain_string := buildInfo[3].(string)
	if domain_string != "" {
		build.Domains = strings.Split(domain_string, ",")
	}
	build.Status = buildInfo[4].(string)

	return &build, nil
}

func createBuild(id string, branchName string, domains []*Domain) (*Build, error) {
	if domains == nil || len(domains) == 0 {
		return nil, errors.New("Empty domain list")
	}

	time := time.Now().Format("2006-01-02 15:04:05")
	domainNames := []string{}
	for _, domain := range domains {
		domainNames = append(domainNames, domain.Domain)
	}

	build := Build{
		BuildId: id,
		Branch:  branchName,
		Domains: domainNames,
		Date:    time,
		Status:  utils.StatusQueued,
	}

	err := saveBuild(&build)
	return &build, err
}

func saveBuild(build *Build) error {
	newBuild := map[string]interface{}{
		"id":      build.BuildId,
		"branch":  build.Branch,
		"date":    build.Date,
		"domains": strings.Join(build.Domains, ","),
		"status":  build.Status,
	}
	buildKey := buildNamespace + build.BuildId
	utils.RedisClient.HMSet(buildKey, newBuild)
	err := utils.RedisClient.Expire(buildKey, buildExpiration).Err()
	return err
}

func updateBuild(id string, status string) error {
	buildKey := buildNamespace + id
	buildInfo, errFetch := fetchBuildInfo(buildKey)
	log.Printf("Build Info: %v", *buildInfo)
	log.Printf("Update Build: %s", status)
	if errFetch != nil {
		return errFetch
	}

	time := time.Now().Format("2006-01-02 15:04:05")
	updatedBuild := map[string]interface{}{
		"id":      id,
		"branch":  buildInfo.Branch,
		"date":    time,
		"domains": strings.Join(buildInfo.Domains, ","),
		"status":  status,
	}
	err := utils.RedisClient.HMSet(buildKey, updatedBuild).Err()
	return err
}

func completeBuild(id string, comm *utils.BranchComm) error {
	responseStatus := utils.CheckStatusList(*comm.BuildStatus)
	log.Printf("Build is complete %v", responseStatus)
	return updateBuild(id, responseStatus)
}

func buildStatusFilter(filter string) []string {
	statusList := []string{}
	switch filter {
	case FilterQueued:
		statusList = append(statusList, utils.StatusQueued)
	case FilterRunning:
		statusList = append(statusList, utils.StatusStarted)
		statusList = append(statusList, utils.StatusBuildCompleted)
	case FilterCompleted:
		statusList = append(statusList, utils.StatusBuildError)
		statusList = append(statusList, utils.StatusDeployError)
		statusList = append(statusList, utils.StatusDeploySuccessful)
	case FilterSuccessful:
		statusList = append(statusList, utils.StatusDeploySuccessful)
	case FilterFailed:
		statusList = append(statusList, utils.StatusBuildError)
		statusList = append(statusList, utils.StatusDeployError)
	}
	return statusList
}

func matchBuildStatus(statusFilterList []string, build *Build) bool {
	if len(statusFilterList) == 0 {
		return true
	}

	for _, value := range statusFilterList {
		if value == build.Status {
			return true
		}
	}

	return false
}
