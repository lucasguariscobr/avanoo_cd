package deploy

import (
	"github.com/avanoo/avanoo_cd/utils"
	"net/http"
	"testing"
)

func TestListDomains(t *testing.T) {
	cleanRedis()
	domains := []Domain{
		Domain{Domain: "pre.avanoo.com", Branch: "development", Host: "pre", ExtraVars: map[string]string{"DOMAIN_NAME": "pre", "HOST_PASSENGER_PORT": "4004"}},
	}
	for _, domain := range domains {
		saveDomain(&domain)
	}
	response := createTestRequest(&JSONTestRequest{
		method:      "GET",
		url:         "/domains",
		handlerFunc: ListDomains,
	})
	status := response.Code
	statusErr := assertStatusCode(status, 200)
	if statusErr != nil {
		t.Error(statusErr.Error())
	}
	jsonResponse := []Domain{}
	decodeTestJSONResponse(response.Body, &jsonResponse)
	diff := testEqualModel(jsonResponse, domains, Domain{})
	if diff != "" {
		t.Error(diff)
	}
}

func TestManageDomains(t *testing.T) {
	cases := map[string]struct {
		domains []Domain
		status  int
	}{
		"create": {
			domains: []Domain{
				{Domain: "pre.avanoo.com", Branch: "development", Host: "pre", ExtraVars: map[string]string{"DOMAIN_NAME": "pre", "HOST_PASSENGER_PORT": "4004"}},
			},
		},
		"update": {
			domains: []Domain{
				{Domain: "prodtest.avanoo.com", Branch: "development", Host: "prodtest", ExtraVars: map[string]string{"DOMAIN_NAME": "prodtest"}},
				{Domain: "prodtest.avanoo.com", Branch: "production", Host: "prodtest", ExtraVars: map[string]string{"DOMAIN_NAME": "prodtest", "EXTRA_APP_REDIS_DATABASE": "2"}},
			},
		},
	}

	for _, value := range cases {
		cleanRedis()
		for _, domain := range value.domains {
			body := encodeMessage(domain)
			createTestRequest(&JSONTestRequest{
				method:      "POST",
				url:         "/domain",
				body:        body,
				handlerFunc: ManageDomain,
			})
		}
		expectedDomain := value.domains[len(value.domains)-1]
		domainKey := buildDomainKey(expectedDomain.Domain)
		domain, _ := fetchDomainInfo(domainKey)
		diff := testEqualModel(*domain, expectedDomain, Domain{})
		if diff != "" {
			t.Error(diff)
		}
	}
}

func TestUpdateDomainBranch(t *testing.T) {
	cases := map[string]struct {
		domain       Domain
		updateDomain string
		updateBranch string
		status       int
	}{
		"update": {
			domain:       Domain{Domain: "pre.avanoo.com", Branch: "development", Host: "pre", ExtraVars: map[string]string{"DOMAIN_NAME": "pre", "HOST_PASSENGER_PORT": "4004"}},
			updateBranch: "production",
			status:       204,
		},
		"unknown": {
			domain:       Domain{Domain: "prodtest.avanoo.com", Branch: "production", Host: "prodtest", ExtraVars: map[string]string{"DOMAIN_NAME": "prodtest", "EXTRA_APP_REDIS_DATABASE": "2"}},
			updateDomain: "pre.avanoo.com",
			status:       400,
		},
	}

	utils.MockCommandDefault()
	utils.MockEmail()

	for _, value := range cases {
		cleanRedis()
		saveDomain(&value.domain)
		updateDomain := getSearchValueWithDefault(value.updateDomain, &value.domain, "Domain")
		updateBranch := getSearchValueWithDefault(value.updateBranch, &value.domain, "Branch")
		body := encodeMessage(&UpdateDomain{
			Domain: updateDomain,
			Branch: updateBranch,
		})
		response := createTestRequest(&JSONTestRequest{
			method:      "POST",
			url:         "/updateDomainBranch",
			body:        body,
			handlerFunc: UpdateDomainBranch,
		})
		status := response.Code
		statusErr := assertStatusCode(status, value.status)
		if statusErr != nil {
			t.Error(statusErr.Error())
		}
		webhookWG.Wait()
		if successTest(status) {
			value.domain.Branch = value.updateBranch
			domainKey := buildDomainKey(updateDomain)
			domain, _ := fetchDomainInfo(domainKey)
			diff := testEqualModel(*domain, value.domain, Domain{})
			if diff != "" {
				t.Error(diff)
			}
		}
	}
}

func TestDetailDomain(t *testing.T) {
	cases := map[string]struct {
		domain       Domain
		searchDomain string
		status       int
	}{
		"exist": {
			domain: Domain{Domain: "pre.avanoo.com", Branch: "development", Host: "pre", ExtraVars: map[string]string{"DOMAIN_NAME": "pre", "HOST_PASSENGER_PORT": "4004"}},
			status: 200,
		},
		"unknown": {
			domain:       Domain{Domain: "prodtest.avanoo.com", Branch: "production", Host: "prodtest", ExtraVars: map[string]string{"DOMAIN_NAME": "prodtest", "EXTRA_APP_REDIS_DATABASE": "2"}},
			searchDomain: "pre.avanoo.com",
			status:       400,
		},
	}

	for _, value := range cases {
		cleanRedis()
		saveDomain(&value.domain)
		searchDomain := getSearchValueWithDefault(value.searchDomain, &value.domain, "Domain")
		response := createTestRequest(&JSONTestRequest{
			method:      "GET",
			url:         "/domain",
			routeVars:   map[string]string{"domain": searchDomain},
			handlerFunc: DetailDomain,
		})
		status := response.Code
		statusErr := assertStatusCode(status, value.status)
		if statusErr != nil {
			t.Error(statusErr.Error())
			continue
		}
		if successTest(status) {
			jsonResponse := Domain{}
			decodeTestJSONResponse(response.Body, &jsonResponse)
			diff := testEqualModel(jsonResponse, value.domain, Domain{})
			if diff != "" {
				t.Error(diff)
			}
		}
	}
}

func TestDeleteDomain(t *testing.T) {
	cases := map[string]struct {
		domain       Domain
		searchDomain string
		status       int
	}{
		"delete": {
			domain: Domain{Domain: "pre.avanoo.com", Branch: "development", Host: "pre", ExtraVars: map[string]string{"DOMAIN_NAME": "pre", "HOST_PASSENGER_PORT": "4004"}},
			status: 204,
		},
		"unknown": {
			domain:       Domain{Domain: "prodtest.avanoo.com", Branch: "production", Host: "prodtest", ExtraVars: map[string]string{"DOMAIN_NAME": "prodtest", "EXTRA_APP_REDIS_DATABASE": "2"}},
			searchDomain: "pre.avanoo.com",
			status:       400,
		},
	}

	for _, value := range cases {
		cleanRedis()
		saveDomain(&value.domain)
		searchDomain := getSearchValueWithDefault(value.searchDomain, &value.domain, "Domain")
		response := createTestRequest(&JSONTestRequest{
			method:      "DELETE",
			url:         "/domain",
			routeVars:   map[string]string{"domain": searchDomain},
			handlerFunc: DeleteDomain,
		})
		status := response.Code
		statusErr := assertStatusCode(status, value.status)
		if statusErr != nil {
			t.Error(statusErr.Error())
			continue
		}
		if successTest(status) {
			domainKey := buildDomainKey(value.domain.Domain)
			domain, err := fetchDomainInfo(domainKey)
			if domain != nil || err == nil {
				t.Error("Domain should have been deleted.")
			}
		}
	}
}

func TestUnknownMessage(t *testing.T) {
	wrongCases := map[string]interface{}{
		"wrong": struct{ Message string }{Message: "wrong format for managedomain"},
		"empty": struct{ unexported string }{unexported: "unexported message"},
	}

	operations := []struct {
		method      string
		url         string
		handlerFunc http.HandlerFunc
	}{
		{method: "POST", url: "/domain", handlerFunc: ManageDomain},
		{method: "POST", url: "/updateDomainBranch", handlerFunc: UpdateDomainBranch},
	}

	cleanRedis()
	for _, value := range wrongCases {
		for _, operation := range operations {
			body := encodeMessage(value)
			response := createTestRequest(&JSONTestRequest{
				method:      operation.method,
				url:         operation.url,
				body:        body,
				handlerFunc: operation.handlerFunc,
			})
			status := response.Code
			statusErr := assertStatusCode(status, 400)
			if statusErr != nil {
				t.Error(statusErr.Error())
			}
		}
	}
}

func TestUnknownDomain(t *testing.T) {
	operations := []struct {
		method      string
		url         string
		handlerFunc http.HandlerFunc
	}{
		{method: "DELETE", url: "/domain", handlerFunc: DeleteDomain},
		{method: "GET", url: "/domain", handlerFunc: DetailDomain},
	}

	cleanRedis()
	for _, operation := range operations {
		response := createTestRequest(&JSONTestRequest{
			method:      operation.method,
			url:         operation.url,
			routeVars:   map[string]string{"domain": ""},
			handlerFunc: operation.handlerFunc,
		})
		status := response.Code
		statusErr := assertStatusCode(status, 400)
		if statusErr != nil {
			t.Error(statusErr.Error())
		}
	}
}
