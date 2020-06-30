package deploy

import (
	"avanoo_cd/utils"
	"fmt"
	"io"
	"testing"
	"time"
	"sort"
)

func TestListenWebHookEvent(t *testing.T) {
	cases := map[string]struct {
		message  PushJSON
		status   int
		response []Build
	}{
		"empty": {
			message:  PushJSON{Ref: "refs/heads/nope", Deleted: false},
			status:   204,
			response: []Build{},
		},
		"development": {
			message: PushJSON{Ref: "refs/heads/development", Deleted: false},
			status:  204,
			response: []Build{
				{Branch: "development", DomainNames: []string{"pre-temp.avanoo.com"}},
			},
		},
	}
	for _, value := range cases {
		cleanRedis()
		createFixtureDomains()
		utils.MockCommandDefault()
		utils.MockEmail()

		body := encodeMessage(&value.message)
		testWebHookEndpoint(t, &body, value.status)

		webhookWG.Wait()
		buildList := fetchBuildList()
		diff := testEqualModel(buildList, value.response, Build{}, "BuildId", "Date", "Status")
		if diff != "" {
			t.Error(diff)
		}
	}
}

func TestListenWebHookQueue(t *testing.T) {
	cleanRedis()
	createFixtureDomains()
	utils.MockCommand(
		&utils.CommandMockReturn{},
		&utils.CommandMockReturn{
			ProcessTime: 2,
			Byte:        nil,
			Error:       nil,
		},
	)
	utils.MockEmail()
	closeWebHooksFunc()
	queueUpdatesCount := 3
	for i := 0; i < queueUpdatesCount; i++ {
		body := encodeMessage(&PushJSON{Ref: "refs/heads/development", Deleted: false})
		testWebHookEndpoint(t, &body, 204)
		webhookWG.Done()
		time.Sleep(time.Duration(300) * time.Millisecond)
	}

	closeWebHooksFunc = CreateDeployContext()
	StartDeployAgent()
	StartBuildAgent()
	webhookWG.Wait()

	buildList := fetchBuildList()
	fmt.Printf("BuildList: %v\n", buildList)
	diff := testEqualModel(
		buildList,
		[]Build{
			{Branch: "development", DomainNames: []string{"pre-temp.avanoo.com"}, Status: utils.StatusCanceled},
			{Branch: "development", DomainNames: []string{"pre-temp.avanoo.com"}, Status: utils.StatusCanceled},
			{Branch: "development", DomainNames: []string{"pre-temp.avanoo.com"}, Status: utils.StatusDeploySuccessful}},
		Build{},
		"BuildId", "Date")
	if diff != "" {
		t.Error(diff)
	}
}

func TestListenWebHookEventError(t *testing.T) {
	cases := map[string]struct {
		message  PushJSON
		status   int
		response []Build
	}{
		"deleted": {
			message: PushJSON{Ref: "refs/heads/development", Deleted: true},
			status:  204,
		},
		"wrongref": {
			message: PushJSON{Ref: "wrongref/development", Deleted: false},
			status:  400,
		},
	}
	for _, value := range cases {
		cleanRedis()
		body := encodeMessage(&value.message)
		testWebHookEndpoint(t, &body, value.status)
	}
}

func TestListenWebHookEventUnknownMessage(t *testing.T) {
	wrongCases := map[string]interface{}{
		"wrong": struct{ Ref int }{Ref: 0123456},
		"empty": struct{ unexported string }{unexported: "unexported message"},
	}
	for _, value := range wrongCases {
		cleanRedis()
		body := encodeMessage(value)
		testWebHookEndpoint(t, &body, 400)
	}
}

func testWebHookEndpoint(t *testing.T, body *io.Reader, expectedStatus int) {
	response := createTestRequest(&JSONTestRequest{
		method:      "POST",
		url:         "/webhook",
		body:        *body,
		handlerFunc: ListenWebHookEvent,
	})
	status := response.Code
	statusErr := assertStatusCode(status, expectedStatus)
	if statusErr != nil {
		t.Error(statusErr.Error())
	}
}

func fetchBuildList() []Build {
	builds, _ := scanBuilds()
	buildList := []Build{}
	for _, build := range builds {
		buildList = append(buildList, *build)
	}
	sort.SliceStable(buildList,
		func(i, j int) bool {
			if buildList[i].Date < buildList[j].Date {
				return true
			}
			if buildList[i].Date > buildList[j].Date {
				return false
			}
			return buildList[i].BuildId < buildList[j].BuildId
		},
	)
	return buildList
}

func createFixtureDomains() {
	domainList := []Domain{
		{Domain: "pre-temp.avanoo.com", Branch: "development", Host: "test"},
		{Domain: "pre-temp2.avanoo.com", Branch: "new-branch", Host: "test"},
		{Domain: "pre-temp3.avanoo.com", Branch: "other-branch", Host: "test"},
	}
	for _, domain := range domainList {
		saveDomain(&domain)
	}
}
