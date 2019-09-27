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
