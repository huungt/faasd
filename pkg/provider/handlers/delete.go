package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/openfaas/faas-provider/types"
	"github.com/openfaas/faasd/pkg"
)

func MakeDeleteHandler(client *http.Client, token string) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Body == nil {
			http.Error(w, "expected a body", http.StatusBadRequest)
			return
		}

		defer r.Body.Close()

		body, _ := io.ReadAll(r.Body)

		req := types.DeleteFunctionRequest{}
		if err := json.Unmarshal(body, &req); err != nil {
			log.Printf("[Delete] error parsing input: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		// namespace moved from the querystring into the body
		namespace := req.Namespace
		if namespace == "" {
			namespace = pkg.DefaultFunctionNamespace
		}

		name := req.FunctionName

		application, err := GetApplication(client, token, name, namespace)
		if err != nil {
			msg := fmt.Sprintf("function %s.%s not found", name, namespace)
			log.Printf("[Delete] %s\n", msg)
			http.Error(w, msg, http.StatusNotFound)
			return
		}

		apiUrl := fmt.Sprintf("http://192.168.0.164:10000/api/application/%s", application.ApplicationID)
		emptyBody := []byte(`{}`)
		request, err := http.NewRequest("DELETE", apiUrl, bytes.NewBuffer(emptyBody))
		request.Header.Set("accept", "application/json")
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+token)

		response, err := client.Do(request)
		if err != nil {
			return
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			log.Printf("[Delete] error removing function %s: %s\n", name, err)
			return
		}

		log.Printf("[Delete] Removed: %s.%s\n", name, namespace)
	}
}
