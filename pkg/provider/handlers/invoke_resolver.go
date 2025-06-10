package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	faasd "github.com/openfaas/faasd/pkg"
)

const watchdogPort = 8080

type InvokeResolver struct {
	client *http.Client
	token  string
}

func NewInvokeResolver(client *http.Client, token string) *InvokeResolver {
	return &InvokeResolver{client: client, token: token}
}

func (i *InvokeResolver) Resolve(functionName string) (url.URL, error) {
	actualFunctionName := functionName
	log.Printf("Resolve: %q\n", actualFunctionName)

	namespace := getNamespaceOrDefault(functionName, faasd.DefaultFunctionNamespace)

	if strings.Contains(functionName, ".") {
		actualFunctionName = strings.TrimSuffix(functionName, "."+namespace)
	}

	function, err := GetFunction(i.client, i.token, actualFunctionName, namespace)
	if err != nil {
		return url.URL{}, fmt.Errorf("%s not found", actualFunctionName)
	}

	serviceIP := function.IP

	urlStr := fmt.Sprintf("http://%s:%d", serviceIP, watchdogPort)

	urlRes, err := url.Parse(urlStr)
	if err != nil {
		return url.URL{}, err
	}

	return *urlRes, nil
}

func getNamespaceOrDefault(name, defaultNamespace string) string {
	namespace := defaultNamespace
	if strings.Contains(name, ".") {
		namespace = name[strings.LastIndexAny(name, ".")+1:]
	}
	return namespace
}
