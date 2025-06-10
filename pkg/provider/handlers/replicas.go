package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/openfaas/faas-provider/types"
)

func MakeReplicaReaderHandler(client *http.Client, token string) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		functionName := vars["name"]
		lookupNamespace := getRequestNamespace(readNamespaceFromQuery(r))

		if f, err := GetFunction(client, token, functionName, lookupNamespace); err == nil {
			found := types.FunctionStatus{
				Name:              functionName,
				Image:             f.image,
				AvailableReplicas: uint64(f.replicas),
				Replicas:          uint64(f.replicas),
				Namespace:         f.namespace,
				Labels:            &f.labels,
				Annotations:       &f.annotations,
				Secrets:           f.secrets,
				EnvVars:           f.envVars,
				EnvProcess:        f.envProcess,
				CreatedAt:         f.createdAt,
			}

			functionBytes, _ := json.Marshal(found)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(functionBytes)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}
