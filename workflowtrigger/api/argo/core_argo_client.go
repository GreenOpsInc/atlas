package argo

import (
	"fmt"
	"github.com/argoproj/argo-cd/pkg/apiclient"
	grpcutil "github.com/argoproj/argo-cd/util/grpc"
	"os"
)

type ArgoApi interface {
	GetClusterApi() ArgoClusterApi
	GetAuthenticatorApi() ArgoAuthenticatorApi
}

type ArgoApiImpl struct {
	apiServerAddress string
	tls              bool
	rawClient        apiclient.Client
	configuredClient apiclient.Client
}

func New() ArgoApi {
	apiServerAddress := getClientCreationData()
	argoClient, tls := getRawArgoClient(apiServerAddress)

	var client ArgoApi
	client = &ArgoApiImpl{apiServerAddress: apiServerAddress, tls: tls, rawClient: argoClient, configuredClient: nil}
	return client
}

func getRawArgoClient(apiServerAddress string) (apiclient.Client, bool) {
	tlsTestResult, err := grpcutil.TestTLS(apiServerAddress)
	if err != nil {
		panic(fmt.Sprintf("Error when testing TLS: %s", err))
	}

	argoClient, err := apiclient.NewClient(
		&apiclient.ClientOptions{
			ServerAddr: apiServerAddress,
			Insecure:   true,
			PlainText:  !tlsTestResult.TLS,
		})
	if err != nil {
		panic(fmt.Sprintf("Error when making new API client: %s", err))
	}
	return argoClient, !tlsTestResult.TLS
}

func (a *ArgoApiImpl) getConfiguredArgoClient(token string) {
	argoClient, err := apiclient.NewClient(&apiclient.ClientOptions{
		Insecure:   true,
		ServerAddr: a.apiServerAddress,
		AuthToken:  token,
		PlainText:  a.tls,
	})
	if err != nil {
		panic(fmt.Sprintf("Error when making properly authenticated client: %s", err))
	}
	a.configuredClient = argoClient
}

func (a *ArgoApiImpl) GetClusterApi() ArgoClusterApi {
	return a
}

func (a *ArgoApiImpl) GetAuthenticatorApi() ArgoAuthenticatorApi {
	return a
}

const (
	UserAccountEnvVar         string = "ARGOCD_USER_ACCOUNT"
	UserAccountPasswordEnvVar string = "ARGOCD_USER_PASSWORD"
	DefaultApiServerAddress   string = "argocd-server.argocd.svc.cluster.local"
	DefaultUserAccount        string = "admin"
)

func getClientCreationData() string {
	argoCdServer := os.Getenv(apiclient.EnvArgoCDServer)
	if argoCdServer == "" {
		argoCdServer = DefaultApiServerAddress
	}

	return argoCdServer
}
