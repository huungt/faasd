package handlers

import (
	"net/http"
	"strings"
)

// GetFunction returns a function that matches name
func GetFunction(client *http.Client, token string, name string, namespace string) (Function, error) {
	fn := Function{}

	mircoservice, err := GetMicroservice(client, token, name, namespace)
	if err != nil {
		return fn, err
	}

	fn.name = mircoservice.MicroserviceName
	fn.namespace = getRequestNamespace(mircoservice.MicroserviceNamespace)
	fn.image = mircoservice.Code
	cmd := strings.Join(mircoservice.Cmd, " ")
	fn.envProcess = cmd
	fn.memoryLimit = int64(mircoservice.Memory)
	fn.replicas = len(mircoservice.InstanceList)

	return fn, nil
}
