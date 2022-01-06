package argodriver

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/argoproj/argo-cd/pkg/apiclient"
	sessionpkg "github.com/argoproj/argo-cd/pkg/apiclient/session"
	grpcutil "github.com/argoproj/argo-cd/util/grpc"
	"github.com/argoproj/argo-cd/util/io"
	"github.com/greenopsinc/util/config"
	"github.com/greenopsinc/util/tlsmanager"
	"greenops.io/client/k8sdriver"
)

const (
	ArgoCDTLSCertPathSuffix string = "/argocd/cert.tls"
)

func (a *ArgoClientDriver) initArgoDriver(userAccount string, userPassword string) error {
	tlsTestResult, err := grpcutil.TestTLS(a.apiServerAddress)
	if err != nil {
		log.Fatalf("Error when testing TLS: %s", err)
	}

	tlsCertPath, err := a.initArgoTLSCert()
	if err != nil {
		log.Println(err.Error())
	}

	token, err := a.generateArgoToken(userAccount, userPassword)
	if err != nil {
		return err
	}

	a.tlsEnabled = tlsTestResult.TLS
	a.tlsCertPath = tlsCertPath
	argoClient, err := apiclient.NewClient(a.getAPIClientOptions(token))
	if err != nil {
		log.Fatalf("Error when making new API client: %s", err)
	}
	a.client = argoClient

	if err = a.watchArgoTLSUpdates(); err != nil {
		log.Fatal("failed to watch argocd tls secret: ", err)
	}
	return nil
}

func (a *ArgoClientDriver) generateArgoToken(userAccount string, password string) (string, error) {
	closer, sessionClient, err := a.client.NewSessionClient()
	if err != nil {
		log.Printf("Error when making new session client: %s", err)
		return "", err
	}
	defer io.Close(closer)

	sessionResponse, err := sessionClient.Create(context.TODO(), &sessionpkg.SessionCreateRequest{Username: userAccount, Password: password})
	if err != nil {
		log.Printf("Error when fetching access token: %s", err)
		return "", err
	}
	return sessionResponse.Token, nil
}

func getClientCreationData(kubernetesDriver *k8sdriver.KubernetesClientNamespaceSecretRestricted) (string, string, string, string) {
	argoCdServer := os.Getenv(apiclient.EnvArgoCDServer)
	if argoCdServer == "" {
		argoCdServer = DefaultApiServerAddress
	}
	argoCdUser := os.Getenv(UserAccountEnvVar)
	if argoCdUser == "" {
		argoCdUser = DefaultUserAccount
	}
	argoCdUserPassword := os.Getenv(UserAccountPasswordEnvVar)
	if argoCdUserPassword == "" {
		secretData := (*kubernetesDriver).GetSecret("argocd-initial-admin-secret", "argocd")
		if secretData != nil {
			argoCdUserPassword = string(secretData["password"])
		}
	}
	argoCdUserToken := "" //os.Getenv(apiclient.EnvArgoCDAuthToken)
	//if argoCdUserToken == "" {
	//	panic("An acces token has to be entered. Not implemented yet.")
	//}
	return argoCdServer, argoCdUser, argoCdUserPassword, argoCdUserToken
}

func (a *ArgoClientDriver) initArgoTLSCert() (string, error) {
	certPEM, err := a.tm.GetClientCertPEM(tlsmanager.ClientArgoCDRepoServer)
	if err != nil {
		log.Println("failed to get argocd certificate from secrets: ", err.Error())
		return "", nil
	}

	confPath, err := config.GetConfigPath()
	if err != nil {
		return "", err
	}
	argoTLSCertPath := fmt.Sprintf("%s/%s", confPath, ArgoCDTLSCertPathSuffix)

	data, err := config.ReadDataFromConfigFile(argoTLSCertPath)
	if err == nil && bytes.Equal(data, certPEM) {
		return argoTLSCertPath, nil
	}

	if err = config.WriteDataToConfigFile(certPEM, argoTLSCertPath); err != nil {
		return "", err
	}
	return argoTLSCertPath, nil
}

func (a *ArgoClientDriver) getAPIClientOptions(token string) *apiclient.ClientOptions {
	options := &apiclient.ClientOptions{
		ServerAddr: a.apiServerAddress,
	}
	if a.tlsCertPath == "" {
		options.Insecure = true
		options.PlainText = !a.tlsEnabled
	} else {
		options.Insecure = false
		options.PlainText = false
		options.CertFile = a.tlsCertPath
	}
	if token != "" {
		options.AuthToken = token
	}
	return options
}

func (a *ArgoClientDriver) watchArgoTLSUpdates() error {
	err := a.tm.WatchClientTLSPEM(tlsmanager.ClientArgoCDRepoServer, func(certPEM []byte, err error) {
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

func (a *ArgoClientDriver) updateArgoTLSCert(certPEM []byte) error {
	confPath, err := config.GetConfigPath()
	if err != nil {
		return err
	}
	argoTLSCertPath := fmt.Sprintf("%s/%s", confPath, ArgoCDTLSCertPathSuffix)

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
