package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/openfaas/faas-provider/types"
)

func MakeUpdateHandler(client *http.Client, token string) func(w http.ResponseWriter, r *http.Request) {

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
			log.Printf("[Update] error parsing input: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		name := req.Service
		namespace := getRequestNamespace(req.Namespace)

		_, err = GetFunction(client, token, name, namespace)
		if err != nil {
			msg := fmt.Sprintf("function: %s.%s not found", name, namespace)
			log.Printf("[Update] %s\n", msg)
			http.Error(w, msg, http.StatusNotFound)
			return
		}

		deploymentDescriptor := buildDeploymentDescriptor(req)

		err = update(client, token, deploymentDescriptor)
		if err != nil {
			log.Printf("[Update] - error deploy descriptor: %s", err)
		}
	}
}

func update(client *http.Client, token string, descriptor SlaDeploymentDescriptor) error {
	data, err := json.Marshal(descriptor)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("PUT", "http://192.168.0.164:10000/api/application/", bytes.NewReader(data))
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
