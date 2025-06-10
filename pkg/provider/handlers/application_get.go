package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

func GetApplications(client *http.Client, token string) ([]Application, error) {
	request, err := http.NewRequest("GET", "http://192.168.0.164:10000/api/applications/", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("accept", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	applications := []Application{}
	err = json.Unmarshal(body, &applications)
	if err != nil {
		return nil, err
	}
	return applications, nil
}

func GetMicroservices(client *http.Client, token string) ([]Microservice, error) {
	request, err := http.NewRequest("GET", "http://192.168.0.164:10000/api/services/", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("accept", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	microservices := []Microservice{}
	err = json.Unmarshal(body, &microservices)
	if err != nil {
		return nil, err
	}
	return microservices, nil
}

func GetApplication(client *http.Client, token string, name string, namespace string) (*Application, error) {
	applications, err := GetApplications(client, token)
	if err != nil {
		return nil, err
	}

	for _, a := range applications {
		if a.ApplicationNamespace == name && a.ApplicationNamespace == namespace {
			return &a, nil
		}
	}
	return nil, errors.Errorf("Application not found")
}

func GetMicroservice(client *http.Client, token string, name string, namespace string) (*Microservice, error) {
	services, err := GetMicroservices(client, token)
	if err != nil {
		return nil, err
	}

	for _, s := range services {
		if s.MicroserviceName == name && s.MicroserviceNamespace == namespace {
			return &s, nil
		}
	}
	return nil, errors.Errorf("Microservice not found")
}
