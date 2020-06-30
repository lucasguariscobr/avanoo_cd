package deploy

import (
	"avanoo_cd/utils"
	"errors"
	"testing"
)

func TestBuildImageError(t *testing.T) {
	cleanRedis()
	utils.MockCommand(
		&utils.CommandMockReturn{
			Byte:  nil,
			Error: errors.New("something went wrong"),
		},
		&utils.CommandMockReturn{},
	)
	utils.MockEmail()

	webhookWG.Add(1)
	releaseBranch("development", []*Domain{&Domain{Domain: "pre-temp.avanoo.com", Branch: "development", Host: "test"}})
	webhookWG.Wait()

	responseBuild := fetchBuildList()[0]
	if responseBuild.Status != utils.StatusBuildError {
		t.Error("Build error not tracked")
	}
}

func TestDomainEnvironments(t *testing.T) {
	cases := map[string]struct {
		domains       	[]*Domain
		build			Build
		expected 		[]string
	}{
		"all": {
			domains: []*Domain{
				{Domain: "pre.avanoo.com", Branch: "development", Host: "pre", ExtraVars: map[string]string{}},
				{Domain: "prodtest.avanoo.com", Branch: "development", Host: "prodtest", ExtraVars: map[string]string{}},
			},
			build: Build{BuildId: "1", Branch: "development", DomainNames: []string{"pre.avanoo.com", "prodtest.avanoo.com"}, Status: utils.StatusQueued},
			expected: []string{"stage", "production"},
		},
		"pre": {
			domains: []*Domain{
				{Domain: "pre.avanoo.com", Branch: "development", Host: "pre", ExtraVars: map[string]string{}},
			},
			build: Build{BuildId: "1", Branch: "development", DomainNames: []string{"pre.avanoo.com"}, Status: utils.StatusQueued},
			expected: []string{"stage"},
		},
		"prod": {
			domains: []*Domain{
				{Domain: "prodtest.avanoo.com", Branch: "development", Host: "prodtest", ExtraVars: map[string]string{}},
			},
			build: Build{BuildId: "1", Branch: "development", DomainNames: []string{"prodtest.avanoo.com"}, Status: utils.StatusQueued},
			expected: []string{"production"},
		},
	}
	for _, value := range cases {
		cleanRedis()
		value.build.Domains = value.domains
		environments := getDomainEnvironments(&value.build)
		diff := testEqualObject(environments, value.expected)
		if diff != "" {
			t.Error(diff)
		}
	}

}
