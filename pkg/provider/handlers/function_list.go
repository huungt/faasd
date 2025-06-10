package handlers

import (
	"net/http"

	"time"
)

type Function struct {
	name        string
	namespace   string
	image       string
	pid         uint32
	replicas    int
	IP          string
	labels      map[string]string
	annotations map[string]string
	secrets     []string
	envVars     map[string]string
	envProcess  string
	memoryLimit int64
	createdAt   time.Time
}

// ListFunctions returns a map of all functions with running tasks on namespace
func ListFunctions(client *http.Client, token string, namespace string) (map[string]*Function, error) {
	functions := make(map[string]*Function)

	services, err := GetMicroservices(client, token)
	if err != nil {
		return functions, err
	}
	for _, service := range services {
		if service.MicroserviceNamespace == namespace {
			function, err := GetFunction(client, token, service.MicroserviceName, service.MicroserviceNamespace)
			if err != nil {
				return functions, err
			}
			functions[service.MicroserviceName] = &function
		}
	}
	return functions, nil
}
