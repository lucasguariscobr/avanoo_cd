package deploy

import (
	"github.com/avanoo/avanoo_cd/utils"
	"fmt"
	"os"
	"testing"
	"time"
)

var closeWebHooksFunc func()

func TestMain(m *testing.M) {
	var closeFunc func()
	closeFunc = utils.ReadConfig()
	closeWebHooksFunc = CreateDeployContext()
	StartDeployAgent()
	StartBuildAgent()
	result := m.Run()
	closeWebHooksFunc()
	closeFunc()
	os.Exit(result)
}

func TestListBuilds(t *testing.T) {
	cases := map[string]struct {
		builds []Build
		status int
	}{
		"sucess": {
			builds: []Build{
				{BuildId: "1", Branch: "development", DomainNames: []string{"pre.avanoo.com"}, Status: utils.StatusQueued},
			},
			status: 200,
		},
	}
	for _, value := range cases {
		cleanRedis()
		for _, build := range value.builds {
			saveBuild(&build)
		}
		response := createTestRequest(&JSONTestRequest{
			method:      "GET",
			url:         "/builds",
			handlerFunc: ListBuilds,
		})
		status := response.Code
		statusErr := assertStatusCode(status, value.status)
		if statusErr != nil {
			t.Error(statusErr.Error())
			continue
		}
		if successTest(status) {
			jsonResponse := []Build{}
			decodeTestJSONResponse(response.Body, &jsonResponse)
			diff := testEqualModel(jsonResponse, value.builds, Build{}, "Date")
			if diff != "" {
				t.Error(diff)
			}
		}
	}
}

func TestListBuildsWithFilter(t *testing.T) {
	cases := map[string]struct {
		filter   string
		response []Build
	}{
		"queued": {
			filter: FilterQueued,
			response: []Build{
				{BuildId: "1", Branch: "development", DomainNames: []string{"pre.avanoo.com"}, Status: utils.StatusQueued},
			},
		},
		"running": {
			filter: FilterRunning,
			response: []Build{
				{BuildId: "2", Branch: "development", DomainNames: []string{"pre.avanoo.com"}, Status: utils.StatusStarted},
				{BuildId: "3", Branch: "development", DomainNames: []string{"pre.avanoo.com"}, Status: utils.StatusBuildCompleted},
			},
		},
		"canceled": {
			filter: FilterCanceled,
			response: []Build{
				{BuildId: "7", Branch: "development", DomainNames: []string{"pre.avanoo.com"}, Status: utils.StatusCanceled},
			},
		},
		"completed": {
			filter: FilterCompleted,
			response: []Build{
				{BuildId: "4", Branch: "development", DomainNames: []string{"pre.avanoo.com"}, Status: utils.StatusBuildError},
				{BuildId: "5", Branch: "development", DomainNames: []string{"pre.avanoo.com"}, Status: utils.StatusDeploySuccessful},
				{BuildId: "6", Branch: "development", DomainNames: []string{"pre.avanoo.com"}, Status: utils.StatusDeployError},
			},
		},
		"successful": {
			filter: FilterSuccessful,
			response: []Build{
				{BuildId: "5", Branch: "development", DomainNames: []string{"pre.avanoo.com"}, Status: utils.StatusDeploySuccessful},
			},
		},
		"failed": {
			filter: FilterFailed,
			response: []Build{
				{BuildId: "4", Branch: "development", DomainNames: []string{"pre.avanoo.com"}, Status: utils.StatusBuildError},
				{BuildId: "6", Branch: "development", DomainNames: []string{"pre.avanoo.com"}, Status: utils.StatusDeployError},
			},
		},
	}
	cleanRedis()
	fixtures := []Build{
		{BuildId: "1", Branch: "development", DomainNames: []string{"pre.avanoo.com"}, Status: utils.StatusQueued},
		{BuildId: "2", Branch: "development", DomainNames: []string{"pre.avanoo.com"}, Status: utils.StatusStarted},
		{BuildId: "3", Branch: "development", DomainNames: []string{"pre.avanoo.com"}, Status: utils.StatusBuildCompleted},
		{BuildId: "4", Branch: "development", DomainNames: []string{"pre.avanoo.com"}, Status: utils.StatusBuildError},
		{BuildId: "5", Branch: "development", DomainNames: []string{"pre.avanoo.com"}, Status: utils.StatusDeploySuccessful},
		{BuildId: "6", Branch: "development", DomainNames: []string{"pre.avanoo.com"}, Status: utils.StatusDeployError},
		{BuildId: "7", Branch: "development", DomainNames: []string{"pre.avanoo.com"}, Status: utils.StatusCanceled},
	}
	for _, fixture := range fixtures {
		time := time.Now().Format("2006-01-02 15:04:05")
		fixture.Date = time
		saveBuild(&fixture)
	}

	for _, value := range cases {
		response := createTestRequest(&JSONTestRequest{
			method:      "GET",
			url:         "/builds",
			queryParams: map[string]string{"filter": value.filter},
			handlerFunc: ListBuilds,
		})
		jsonResponse := []Build{}
		decodeTestJSONResponse(response.Body, &jsonResponse)
		diff := testEqualModel(jsonResponse, value.response, Build{}, "Date")
		if diff != "" {
			t.Error(diff)
		}
	}
}

func TestCreateBuild(t *testing.T) {
	cases := map[string]struct {
		id       string
		branch   string
		domains  []*Domain
		response Build
	}{
		"success": {
			id:       "1",
			branch:   "development",
			domains:  []*Domain{{Domain: "pre.avanoo.com", Branch: "development"}},
			response: Build{BuildId: "1", Branch: "development", DomainNames: []string{"pre.avanoo.com"}, Status: "Queued"},
		},
		"empty domain list": {
			id:     "1",
			branch: "development",
		},
	}

	for _, value := range cases {
		cleanRedis()
		response, err := createBuild(value.id, value.branch, value.domains)
		if err != nil {
			if value.response.BuildId != "" {
				t.Error("Not empty error", err.Error())
			}
		} else {
			diff := testEqualModel(*response, value.response, Build{}, "Date", "Domains")
			if diff != "" {
				t.Error(diff)
			}
		}
	}

}

func TestUpdateBuild(t *testing.T) {
	cases := map[string]struct {
		build   Build
		updates []string
		err     error
	}{
		"success": {
			build:   Build{BuildId: "1", Branch: "development", DomainNames: []string{"pre.avanoo.com"}, Status: utils.StatusQueued},
			updates: []string{utils.StatusStarted, utils.StatusBuildError},
			err:     nil,
		},
	}

	var err error
	for _, value := range cases {
		cleanRedis()
		saveBuild(&value.build)
		for _, update := range value.updates {
			err = updateBuild(&value.build, update)
		}

		if err != value.err {
			t.Error("Not empty error", err.Error())
		}

		err = testUpdateStatus(value.build.BuildId, value.updates)
		if err != nil {
			t.Error(err)
		}
	}

}

func testUpdateStatus(buildId string, updates []string) error {
	lastUpdate := updates[len(updates)-1]
	buildKey := buildNamespace + buildId
	build, _ := fetchBuildInfo(buildKey)
	if build.Status != lastUpdate {
		return fmt.Errorf("Build wasn't updated: %v - %v", build.Status, lastUpdate)
	}
	return nil
}
