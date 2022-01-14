package argo

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/argoproj/argo-cd/pkg/apiclient"
	grpcutil "github.com/argoproj/argo-cd/util/grpc"
	"github.com/greenopsinc/util/config"
	"github.com/greenopsinc/util/tlsmanager"
)

const (
	UserAccountEnvVar         string = "ARGOCD_USER_ACCOUNT"
	UserAccountPasswordEnvVar string = "ARGOCD_USER_PASSWORD"
	DefaultApiServerAddress   string = "argocd-server.argocd.svc.cluster.local"
	DefaultUserAccount        string = "admin"
	ArgoTLSCertPathSuffix     string = "/argocd/cert.tls"
)

type ArgoApi interface {
	GetClusterApi() ArgoClusterApi
	GetAuthenticatorApi() ArgoAuthenticatorApi
}

type ArgoApiImpl struct {
	apiServerAddress string
	tm               tlsmanager.Manager
	tlsEnabled       bool
	tlsCertPath      string
	rawClient        apiclient.Client
	configuredClient apiclient.Client
}

func New(tm tlsmanager.Manager) ArgoApi {
	apiServerAddress := getClientCreationData()
	argoApi := &ArgoApiImpl{apiServerAddress: apiServerAddress, tm: tm}
	argoApi.initArgoClient()
	return argoApi
}

func (a *ArgoApiImpl) GetClusterApi() ArgoClusterApi {
	return a
}

func (a *ArgoApiImpl) GetAuthenticatorApi() ArgoAuthenticatorApi {
	return a
}

func getClientCreationData() string {
	argoCdServer := os.Getenv(apiclient.EnvArgoCDServer)
	if argoCdServer == "" {
		argoCdServer = DefaultApiServerAddress
	}
	return argoCdServer
}

func (a *ArgoApiImpl) initArgoClient() {
	tlsTestResult, err := grpcutil.TestTLS(a.apiServerAddress)
	if err != nil {
		log.Fatalf("Error when testing TLS: %s", err)
	}

	tlsCertPath, err := a.initArgoTLSCert()
	if err != nil {
		log.Println(err.Error())
	}

	a.tlsEnabled = tlsTestResult.TLS
	a.tlsCertPath = tlsCertPath
	argoClient, err := apiclient.NewClient(a.getAPIClientOptions(""))
	if err != nil {
		log.Fatalf("Error when making new API client: %s", err)
	}
	a.rawClient = argoClient

	if err = a.watchArgoTLSUpdates(); err != nil {
		log.Fatal("failed to watch argocd tls secret: ", err)
	}
}

func (a *ArgoApiImpl) initArgoTLSCert() (string, error) {
	certPEM, err := a.tm.GetClientCertPEM(tlsmanager.ClientArgoCDRepoServer)
	log.Println("found argocd cert PEM: ", certPEM)
	if err != nil {
		log.Println("failed to get argocd certificate from secrets: ", err.Error())
		return "", nil
	}
	if certPEM == nil || len(certPEM) == 0 {
		log.Println("argocd cert pem is not found")
		return "", nil
	}

	confPath, err := config.GetConfigPath()
	if err != nil {
		return "", err
	}
	argoTLSCertPath := fmt.Sprintf("%s/%s", confPath, ArgoTLSCertPathSuffix)

	data, err := config.ReadDataFromConfigFile(argoTLSCertPath)
	if err == nil && bytes.Equal(data, certPEM) {
		return argoTLSCertPath, nil
	}

	if err = config.WriteDataToConfigFile(certPEM, argoTLSCertPath); err != nil {
		return "", err
	}
	return argoTLSCertPath, nil
}

func (a *ArgoApiImpl) configureArgoClient(token string) {
	argoClient, err := apiclient.NewClient(a.getAPIClientOptions(token))
	if err != nil {
		log.Fatalf("Error when making properly authenticated client: %s", err)
	}
	a.configuredClient = argoClient
}

func (a *ArgoApiImpl) getAPIClientOptions(token string) *apiclient.ClientOptions {
	options := &apiclient.ClientOptions{
		ServerAddr: a.apiServerAddress,
	}
	if a.tlsCertPath == "" {
		log.Println("getAPIClientOptions: tls certificate is not found, setting insecure tls")
		options.Insecure = true
		options.PlainText = !a.tlsEnabled
	} else {
		log.Println("getAPIClientOptions: tls certificate is found, setting insecure to fase and setting cert path")
		options.Insecure = false
		options.PlainText = false
		options.CertFile = a.tlsCertPath
	}
	if token != "" {
		options.AuthToken = token
	}
	return options
}

func (a *ArgoApiImpl) watchArgoTLSUpdates() error {
	err := a.tm.WatchClientTLSPEM(tlsmanager.ClientArgoCDRepoServer, tlsmanager.NamespaceArgoCD, func(certPEM []byte, err error) {
		log.Printf("in watchArgoTLSUpdates, conf = %v, err = %v\n", certPEM, err)
		if err != nil {
			log.Fatalf("an error occurred in the watch %s client: %s", tlsmanager.ClientArgoCDRepoServer, err.Error())
		}
		if err = a.updateArgoTLSCert(certPEM); err != nil {
			log.Fatal("an error occurred during argocd client tls config update: ", err)
		}
	})
	return err
}

func (a *ArgoApiImpl) updateArgoTLSCert(certPEM []byte) error {
	confPath, err := config.GetConfigPath()
	if err != nil {
		return err
	}
	argoTLSCertPath := fmt.Sprintf("%s/%s", confPath, ArgoTLSCertPathSuffix)

	if certPEM == nil {
		if err = config.DeleteConfigFile(argoTLSCertPath); err != nil {
			return err
		}
	}

	data, err := config.ReadDataFromConfigFile(argoTLSCertPath)
	if err == nil && bytes.Equal(data, certPEM) {
		return nil
	}

	if err = config.WriteDataToConfigFile(certPEM, argoTLSCertPath); err != nil {
		return err
	}
	a.tlsCertPath = argoTLSCertPath
	return nil
}
