package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/openfaas/faas-provider/types"
)

const annotationLabelPrefix = "com.openfaas.annotations."

// MakeDeployHandler returns a handler to deploy a function
func MakeDeployHandler(client *http.Client, token string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Body == nil {
			http.Error(w, "expected a body", http.StatusBadRequest)
			return
		}

		defer r.Body.Close()

		body, _ := io.ReadAll(r.Body)

		req := types.FunctionDeployment{}
		err := json.Unmarshal(body, &req)
		if err != nil {
			log.Printf("[Deploy] - error parsing input: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		deploymentDescriptor := buildDeploymentDescriptor(req)

		err = deploy(client, token, deploymentDescriptor)
		if err != nil {
			log.Printf("[Deploy] - error deploy descriptor: %s", err)
		}

	}
}

func buildDeploymentDescriptor(functionDeployment types.FunctionDeployment) SlaDeploymentDescriptor {
	microservice := Microservice{
		MicroserviceID:        "",
		MicroserviceName:      functionDeployment.Service,
		MicroserviceNamespace: getRequestNamespace(functionDeployment.Namespace),
		Virtualization:        "container",
		Cmd:                   strings.Fields(functionDeployment.EnvProcess),
		Vcpu:                  1,
		Storage:               200,
		Code:                  functionDeployment.Image,
	}

	application := Application{
		ApplicationID:        "",
		ApplicationName:      functionDeployment.Service,
		ApplicationNamespace: getRequestNamespace(functionDeployment.Namespace),
		ApplicationDesc:      "OpenFaaS deployment ",
		Microservices:        []Microservice{microservice},
	}

	descriptor := SlaDeploymentDescriptor{
		SlaVersion:   "v2.0",
		CustomerID:   "Admin",
		Applications: []Application{application},
	}

	return descriptor
}

func deploy(client *http.Client, token string, descriptor SlaDeploymentDescriptor) error {
	data, err := json.Marshal(descriptor)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", "http://192.168.0.164:10000/api/application/", bytes.NewReader(data))
	request.Header.Set("accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return err
	}

	body, _ := io.ReadAll(response.Body)

	application := Application{}
	err = json.Unmarshal(body, &application)
	if err != nil {
		return err
	}

	err = deployService(client, token, application)
	return nil
}

func deployService(client *http.Client, token string, application Application) error {
	for _, service := range application.Microservices {

		apiUrl := fmt.Sprintf("http://192.168.0.164:10000/api/service/%s/instance", service.MicroserviceID)
		emptyBody := []byte(`{}`)
		request, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(emptyBody))
		request.Header.Set("accept", "application/json")
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+token)

		response, err := client.Do(request)
		if err != nil {
			return err
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return err
		}
	}

	return nil
}
