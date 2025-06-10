package config

import (
	"time"

	types "github.com/openfaas/faas-provider/types"
)

type ProviderConfig struct {
	// Sock is the address of the containerd socket
	OakestraAPI          string
	OakestraUser         string
	OakestraPassword     string
	OakestraOrganization string
}

// ReadFromEnv loads the FaaSConfig and the Containerd specific config form the env variables
func ReadFromEnv(hasEnv types.HasEnv) (*types.FaaSConfig, *ProviderConfig, error) {
	config, err := types.ReadConfig{}.Read(hasEnv)
	if err != nil {
		return nil, nil, err
	}

	serviceTimeout := types.ParseIntOrDurationValue(hasEnv.Getenv("service_timeout"), time.Second*60)

	config.ReadTimeout = serviceTimeout
	config.WriteTimeout = serviceTimeout
	config.EnableBasicAuth = false
	config.MaxIdleConns = types.ParseIntValue(hasEnv.Getenv("max_idle_conns"), 1024)
	config.MaxIdleConnsPerHost = types.ParseIntValue(hasEnv.Getenv("max_idle_conns_per_host"), 1024)

	port := types.ParseIntValue(hasEnv.Getenv("port"), 8081)
	config.TCPPort = &port

	providerConfig := &ProviderConfig{
		OakestraAPI:          types.ParseString(hasEnv.Getenv("oakestra_api"), "http://192.168.0.164:10000/api"),
		OakestraUser:         types.ParseString(hasEnv.Getenv("oakestra_user"), "Admin"),
		OakestraPassword:     types.ParseString(hasEnv.Getenv("oakestra_password"), "Admin"),
		OakestraOrganization: types.ParseString(hasEnv.Getenv("oakestra_organization"), "string"),
	}

	return config, providerConfig, nil
}
