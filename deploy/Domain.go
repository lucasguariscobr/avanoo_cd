package deploy

import (
	"avanoo_cd/utils"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type UpdateDomain struct {
	Domain string `json:"domain" example:"pre.avanoo.com"`
	Branch string `json:"branch" example:"development"`
}

type Domain struct {
	Domain    string            `json:"domain" example:"pre.avanoo.com"`
	Branch    string            `json:"branch" example:"development"`
	Host      string            `json:"host,omitempty"`
	ExtraVars map[string]string `json:"extra_vars,omitempty"`
}

var domainNamespace = "domain_"

// ManageDomain godoc
// @Summary Manage Avanoo Domains
// @Description Creates or updates avanoo domains.
// @Description When the domain already exists, its content will be replaced by the params provided on the request.
// @Tags domains
// @Accept  json
// @Produce  json
// @Param domain body deploy.Domain true "Domain"
// @Success 204
// @Failure 400 {object} utils.JSONErrror
// @Failure 404
// @Failure 405
// @Router /domain [post]
func ManageDomain(w http.ResponseWriter, r *http.Request) {
	domainJSON, err := utils.DecodeMsg(r, &Domain{}, true)
	if err != nil {
		utils.WriteJSONError(w, err.Error())
		return
	}

	domain := domainJSON.(*Domain)
	if domain.Domain == "" ||
		domain.Branch == "" {
		utils.WriteJSONError(w, "It wasn't possible to create or update this domain")
		return
	}

	err = saveDomain(domain)
	if err != nil {
		utils.WriteJSONError(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateDomainBranch godoc
// @Summary Update a Domain's branch
// @Description Update the branch used by an existing domain.
// @Description It will update the domain after receiving the new branch value.
// @Tags domains
// @Accept  json
// @Produce  json
// @Param domain body deploy.UpdateDomain true "Update Domain Info"
// @Success 204
// @Failure 400 {object} utils.JSONErrror
// @Failure 404
// @Failure 405
// @Router /updateDomainBranch [post]
func UpdateDomainBranch(w http.ResponseWriter, r *http.Request) {
	updateDomainJSON, err := utils.DecodeMsg(r, &UpdateDomain{}, true)
	if err != nil {
		utils.WriteJSONError(w, err.Error())
		return
	}

	updateDomain := updateDomainJSON.(*UpdateDomain)
	if updateDomain.Domain == "" ||
		updateDomain.Branch == "" {
		utils.WriteJSONError(w, "It wasn't possible to update this domain")
		return
	}

	_, errFetch := fetchDomainInfo(domainNamespace + updateDomain.Domain)
	if errFetch != nil {
		log.Printf(errFetch.Error())
		utils.WriteJSONError(w, "It wasn't possible to fetch any info from this domain")
		return
	}

	err = utils.RedisClient.HSet(context.Background(), domainNamespace+updateDomain.Domain, "branch", updateDomain.Branch).Err()
	if err != nil {
		utils.WriteJSONError(w, err.Error())
		return
	}

	webhookWG.Add(1)
	go verifyBranch(updateDomain.Branch)
	w.WriteHeader(http.StatusNoContent)
}

// DeleteDomain godoc
// @Summary Delete an existing domain's definition
// @Description Delete an existing domain's definition.
// @Description This operation does not stop the service or modifies the exisitng domain.
// @Tags domains
// @Param domain path string true "Name of the domain"
// @Success 204
// @Failure 400 {object} utils.JSONErrror
// @Failure 404
// @Failure 405
// @Router /domain/{domain} [delete]
func DeleteDomain(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	domainName, _ := params["domain"]
	if domainName == "" {
		utils.WriteJSONError(w, "Empty domain")
		return
	}

	_, errFetch := fetchDomainInfo(domainNamespace + domainName)
	if errFetch != nil {
		log.Printf(errFetch.Error())
		utils.WriteJSONError(w, "It wasn't possible to fetch any info from this domain")
		return
	}

	err := utils.RedisClient.Del(context.Background(), domainNamespace + domainName).Err()
	if err != nil {
		utils.WriteJSONError(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DetailDomain godoc
// @Summary Return attributes of an existing domain
// @Description Return attributes of an existing domain
// @Tags domains
// @Produce  json
// @Param domain path string true "Name of the domain"
// @Success 200 {object} deploy.Domain true
// @Failure 400 {object} utils.JSONErrror
// @Failure 404
// @Failure 405
// @Router /domain/{domain} [get]
func DetailDomain(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	domainName, _ := params["domain"]
	if domainName == "" {
		utils.WriteJSONError(w, "Empty domain")
		return
	}

	domain, err := fetchDomainInfo(domainNamespace + domainName)
	if err != nil {
		log.Printf(err.Error())
		utils.WriteJSONError(w, "It wasn't possible to fetch any info from this domain")
		return
	}

	json.NewEncoder(w).Encode(*domain)
}

// ListDomains godoc
// @Summary List all domains
// @Description Return all domains and their parameters.
// @Tags domains
// @Produce  json
// @Success 200 {array} deploy.Domain "List of Domains"
// @Failure 400 {object} utils.JSONErrror
// @Failure 404
// @Failure 405
// @Router /domains [get]
func ListDomains(w http.ResponseWriter, r *http.Request) {
	domains, err := scanDomains()
	if err != nil {
		log.Printf(err.Error())
		utils.WriteJSONError(w, "It wasn't possible to fetch the list of domains")
		return
	}

	response := []Domain{}
	for _, value := range domains {
		response = append(response, *value)
	}

	json.NewEncoder(w).Encode(response)
}

func scanDomains() (map[string]*Domain, error) {
	domains := map[string]*Domain{}
	keyIter := utils.RedisClient.Scan(context.Background(), 0, domainNamespace+"*", 100).Iterator()
	for keyIter.Next(context.Background()) {
		domainKey := keyIter.Val()
		domain, err := fetchDomainInfo(domainKey)
		if err != nil {
			return nil, err
		}
		domains[domain.Domain] = domain
	}
	if err := keyIter.Err(); err != nil {
		return nil, err
	}

	return domains, nil
}

func fetchDomainInfo(domainName string) (*Domain, error) {
	domainInfo, errRedis := utils.RedisClient.HGetAll(context.Background(), domainName).Result()
	if errRedis != nil {
		return nil, errRedis
	}
	if len(domainInfo) == 0 {
		return nil, errors.New("no matching domain")
	}

	domain := Domain{
		ExtraVars: map[string]string{},
	}
	if domainInfo != nil {
		for k, v := range domainInfo {
			switch k {
			case "domain":
				domain.Domain = v
			case "branch":
				domain.Branch = v
			case "host":
				domain.Host = v
			default:
				domain.ExtraVars[k] = v
			}
		}
	}

	return &domain, nil
}

func saveDomain(domain *Domain) error {
	domainMap := map[string]interface{}{
		"domain": domain.Domain,
		"branch": domain.Branch,
		"host":   domain.Host,
	}

	if domain.ExtraVars != nil {
		for k, v := range domain.ExtraVars {
			domainMap[k] = v
		}
	}

	domainKey := domainNamespace + domain.Domain
	err := utils.RedisClient.HMSet(context.Background(), domainKey, domainMap).Err()
	return err
}

func findBuildDomains(build *Build) []*Domain {
	if build == nil {
		return nil
	}
	response := []*Domain{}

	domains, _ := scanDomains()
	for _, buildDomain := range build.DomainNames {
		domain, ok := domains[buildDomain]
		if ok {
			response = append(response, domain)
		}
	}

	return response
}

func buildDomainKey(domain string) string {
	return domainNamespace + domain
}
