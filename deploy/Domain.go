package deploy

import (
	"avanoo_cd/utils"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type DomainJSON struct {
	Domain    string            `json:"domain"`
	Branch    string            `json:"branch"`
	Host      string            `json:"host"`
	ExtraVars map[string]string `json:"extra_vars,omitempty"`
}

var domainNamespace = "domain_"

func decodeDomain(r *http.Request) (*DomainJSON, error) {
	if r.Body == nil {
		return nil, errors.New("no request body")
	}

	var domainJSON DomainJSON
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&domainJSON)
	if err != nil {
		return nil, err
	}

	return &domainJSON, nil
}

func ManageDomain(w http.ResponseWriter, r *http.Request) {
	domainJSON, err := decodeDomain(r)
	if err != nil {
		utils.WriteJSONError(w, err.Error())
		return
	}

	newDomain := map[string]interface{}{
		"domain": domainJSON.Domain,
		"branch": domainJSON.Branch,
		"host":   domainJSON.Host,
	}

	if domainJSON.ExtraVars != nil {
		for k, v := range domainJSON.ExtraVars {
			newDomain[k] = v
		}
	}

	err = utils.RedisClient.HMSet(domainNamespace+domainJSON.Domain, newDomain).Err()
	if err != nil {
		utils.WriteJSONError(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func UpdateDomainBranch(w http.ResponseWriter, r *http.Request) {
	domainJSON, err := decodeDomain(r)
	if err != nil {
		utils.WriteJSONError(w, err.Error())
		return
	}

	if domainJSON.Domain == "" ||
		domainJSON.Branch == "" {
		log.Printf(err.Error())
		utils.WriteJSONError(w, "It wasn't possible to update this domain")
		return
	}

	err = utils.RedisClient.HSet(domainNamespace+domainJSON.Domain, "branch", domainJSON.Branch).Err()
	if err != nil {
		utils.WriteJSONError(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteDomain(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	domainName, _ := params["domain"]
	if domainName == "" {
		utils.WriteJSONError(w, "Empty domain")
		return
	}

	err := utils.RedisClient.Del(domainNamespace + domainName).Err()
	if err != nil {
		utils.WriteJSONError(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

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

func ListDomains(w http.ResponseWriter, r *http.Request) {
	domains, err := scanDomains()
	if err != nil {
		log.Printf(err.Error())
		utils.WriteJSONError(w, "It wasn't possible to fetch the list of domains")
		return
	}

	response := []DomainJSON{}
	for _, value := range domains {
		response = append(response, *value)
	}

	json.NewEncoder(w).Encode(response)
}

func scanDomains() (map[string]*DomainJSON, error) {
	domains := map[string]*DomainJSON{}
	keyIter := utils.RedisClient.Scan(0, domainNamespace+"*", 100).Iterator()
	for keyIter.Next() {
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

func fetchDomainInfo(domainName string) (*DomainJSON, error) {
	domainInfo, errRedis := utils.RedisClient.HGetAll(domainName).Result()
	if errRedis != nil {
		return nil, errRedis
	}

	domain := DomainJSON{
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
