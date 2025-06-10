package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	bootstrap "github.com/openfaas/faas-provider"
	"github.com/openfaas/faas-provider/logs"
	"github.com/openfaas/faas-provider/proxy"
	"github.com/openfaas/faas-provider/types"
	faasd "github.com/openfaas/faasd/pkg"
	faasdlogs "github.com/openfaas/faasd/pkg/logs"
	"github.com/openfaas/faasd/pkg/provider/config"
	"github.com/openfaas/faasd/pkg/provider/handlers"
	"github.com/spf13/cobra"
)

const secretDirPermission = 0755

type LoginData struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	Organization string `json:"organization"`
}

type Token struct {
	Refresh_token string `json:"refresh_token"`
	Token         string `json:"token"`
}

func makeProviderCmd() *cobra.Command {
	var command = &cobra.Command{
		Use:   "provider",
		Short: "Run the faasd-provider",
	}

	command.RunE = runProviderE
	command.PreRunE = preRunE

	return command
}

func runProviderE(cmd *cobra.Command, _ []string) error {

	config, providerConfig, err := config.ReadFromEnv(types.OsEnv{})
	if err != nil {
		return err
	}

	log.Printf("faasd-provider starting..\tService Timeout: %s\n", config.WriteTimeout.String())
	printVersion()

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	if err := os.WriteFile(path.Join(wd, "hosts"),
		[]byte(`127.0.0.1	localhost`), workingDirectoryPermission); err != nil {
		return fmt.Errorf("cannot write hosts file: %s", err)
	}

	if err := os.WriteFile(path.Join(wd, "resolv.conf"),
		[]byte(`nameserver 8.8.8.8
nameserver 8.8.4.4`), workingDirectoryPermission); err != nil {
		return fmt.Errorf("cannot write resolv.conf file: %s", err)
	}

	// 	cni, err := cninetwork.InitNetwork()
	// 	if err != nil {
	// 		return err
	// 	}

	client := &http.Client{}

	// token wird für Authentifizierung benötigt
	token, err := getToken(client, providerConfig)
	if err != nil {
		return fmt.Errorf("cannot get token: %s", err)
	}

	invokeResolver := handlers.NewInvokeResolver(client, token)

	baseUserSecretsPath := path.Join(wd, "secrets")
	if err := moveSecretsToDefaultNamespaceSecrets(
		baseUserSecretsPath,
		faasd.DefaultFunctionNamespace); err != nil {
		return err
	}

	bootstrapHandlers := types.FaaSHandlers{
		FunctionProxy:  httpHeaderMiddleware(proxy.NewHandlerFunc(*config, invokeResolver, false)),
		DeleteFunction: httpHeaderMiddleware(handlers.MakeDeleteHandler(client, token)),
		DeployFunction: httpHeaderMiddleware(handlers.MakeDeployHandler(client, token)),
		FunctionLister: httpHeaderMiddleware(handlers.MakeReadHandler(client, token)),
		FunctionStatus: httpHeaderMiddleware(handlers.MakeReplicaReaderHandler(client, token)),
		//ScaleFunction:   httpHeaderMiddleware(handlers.MakeReplicaUpdateHandler(client, token)),
		UpdateFunction: httpHeaderMiddleware(handlers.MakeUpdateHandler(client, token)),
		Health:         httpHeaderMiddleware(func(w http.ResponseWriter, r *http.Request) {}),
		Info:           httpHeaderMiddleware(handlers.MakeInfoHandler(faasd.Version, faasd.GitCommit)),
		//ListNamespaces:  httpHeaderMiddleware(handlers.MakeNamespacesLister(client)),
		//Secrets:         httpHeaderMiddleware(handlers.MakeSecretHandler(client.NamespaceService(), baseUserSecretsPath)),
		Logs: httpHeaderMiddleware(logs.NewLogHandlerFunc(faasdlogs.New(), config.ReadTimeout)),
		//MutateNamespace: httpHeaderMiddleware(handlers.MakeMutateNamespace(client)),
	}

	log.Printf("Listening on: 0.0.0.0:%d", *config.TCPPort)
	bootstrap.Serve(cmd.Context(), &bootstrapHandlers, config)
	return nil
}

/*
* Mutiple namespace support was added after release 0.13.0
* Function will help users to migrate on multiple namespace support of faasd
 */
func moveSecretsToDefaultNamespaceSecrets(baseSecretPath string, defaultNamespace string) error {
	newSecretPath := path.Join(baseSecretPath, defaultNamespace)

	err := ensureSecretsDir(newSecretPath)
	if err != nil {
		return err
	}

	files, err := os.ReadDir(baseSecretPath)
	if err != nil {
		return err
	}

	for _, f := range files {
		if !f.IsDir() {

			newPath := path.Join(newSecretPath, f.Name())

			// A non-nil error means the file wasn't found in the
			// destination path
			if _, err := os.Stat(newPath); err != nil {
				oldPath := path.Join(baseSecretPath, f.Name())

				if err := copyFile(oldPath, newPath); err != nil {
					return err
				}

				log.Printf("[Migration] Copied %s to %s", oldPath, newPath)
			}
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	inputFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening %s failed %w", src, err)
	}
	defer inputFile.Close()

	outputFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_APPEND, secretDirPermission)
	if err != nil {
		return fmt.Errorf("opening %s failed %w", dst, err)
	}
	defer outputFile.Close()

	// Changed from os.Rename due to issue in #201
	if _, err := io.Copy(outputFile, inputFile); err != nil {
		return fmt.Errorf("writing into %s failed %w", outputFile.Name(), err)
	}

	return nil
}

func httpHeaderMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-OpenFaaS-EULA", "openfaas-ce")
		next.ServeHTTP(w, r)
	}
}

func getToken(client *http.Client, providerConfig *config.ProviderConfig) (string, error) {
	loginData := LoginData{
		Username:     providerConfig.OakestraUser,
		Password:     providerConfig.OakestraPassword,
		Organization: providerConfig.OakestraOrganization,
	}

	data, err := json.Marshal(loginData)
	if err != nil {
		return "", err
	}

	request, err := http.NewRequest("POST", providerConfig.OakestraAPI+"/auth/login", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}

	response, err := client.Do(request)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)

	token := Token{}
	err = json.Unmarshal(body, &token)
	if err != nil {
		return "", err
	}

	return token.Token, err
}
