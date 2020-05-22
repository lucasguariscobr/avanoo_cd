package utils

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
)
// Env godoc
// @Summary Download .env file
// @Description Download an .env file for your service
// @Param service query string true "Service Name" Enums(app)
// @Param environment query string true "Environment Name" Enums(development, test, stage, production)
// @Tags env
// @Produce plain
// @Success 200
// @Failure 400 {object} utils.JSONErrror
// @Failure 404
// @Failure 405
// @Router /env [get]
func ConsulTemplate(w http.ResponseWriter, r *http.Request) {
	var commandBuilder strings.Builder

	params := r.URL.Query()
	serviceName := params.Get("service")
	environmentName := params.Get("environment")
	if serviceName == "" {
		WriteJSONError(w, "Empty service")
		return
	}

	if environmentName == "" {
		WriteJSONError(w, "Empty environment")
		return
	}
	
	outputFile := fmt.Sprintf("/tmp/.env.%s.%s", serviceName, environmentName)
	commandBuilder.WriteString(fmt.Sprintf("CONSUL_TEMPLATE_SERVICE=%s ", serviceName))
	commandBuilder.WriteString(fmt.Sprintf("CONSUL_TEMPLATE_ENVIRONMENT=%s ", environmentName))
	commandBuilder.WriteString(fmt.Sprintf("CONSUL_HTTP_TOKEN=%s ", ConsulToken))
	commandBuilder.WriteString(fmt.Sprintf("consul-template -template \"%s/service_env_template.tcl:%s\" -once", Playbooks.DefaultPath, outputFile))

	//log.Printf("Consul Template Command: %s", commandBuilder.String())
	consulTemplateCommand := commandBuilder.String()
	cmd := exec.CommandContext(context.Background(), "/bin/bash", "-c", consulTemplateCommand)
	outByte, err := cmd.CombinedOutput()
	var output strings.Builder
	output.Write(outByte)

	if err != nil {
		log.Printf(output.String())
		WriteJSONError(w, "An error ocurred")
		return
	} 
	http.ServeFile(w, r, outputFile)
}