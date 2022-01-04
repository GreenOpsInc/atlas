package argoauthenticator

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/argoproj/argo-cd/pkg/apiclient"
	"github.com/argoproj/argo-cd/pkg/apiclient/account"
	grpcutil "github.com/argoproj/argo-cd/util/grpc"
	"google.golang.org/grpc/status"
	"greenops.io/workflowtrigger/tlsmanager"
	"greenops.io/workflowtrigger/util/config"
)

type RbacResource string
type RbacAction string

const (
	ClusterResource     RbacResource = "clusters"
	ApplicationResource RbacResource = "applications"

	SyncAction     RbacAction = "sync"
	CreateAction   RbacAction = "create"
	UpdateAction   RbacAction = "update"
	GetAction      RbacAction = "get"
	DeleteAction   RbacAction = "delete"
	ActionAction   RbacAction = "action"
	OverrideAction RbacAction = "override"
)

const (
	UserAccountEnvVar         string = "ARGOCD_USER_ACCOUNT"
	UserAccountPasswordEnvVar string = "ARGOCD_USER_PASSWORD"
	DefaultApiServerAddress   string = "argocd-server.argocd.svc.cluster.local"
	DefaultUserAccount        string = "admin"
	ArgoCDTLSCertPathSuffix   string = "/argocd/cert.tls"
)

type ArgoAuthenticatorApi interface {
	CheckRbacPermissions(action RbacAction, resource RbacResource, subresource string) bool
	Middleware(next http.Handler) http.Handler
}

type argoAuthenticatorApi struct {
	apiServerAddress string
	tm               tlsmanager.Manager
	tlsEnabled       bool
	tlsCertPath      string
	rawClient        apiclient.Client
	configuredClient apiclient.Client
}

func New(tm tlsmanager.Manager) ArgoAuthenticatorApi {
	apiServerAddress := getClientCreationData()
	argoApi := &argoAuthenticatorApi{apiServerAddress: apiServerAddress, tm: tm}
	argoApi.initArgoClient()
	return argoApi
}

func (a *argoAuthenticatorApi) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		splitToken := strings.Split(token, "Bearer ")
		token = splitToken[1]
		a.configureArgoClient(token)

		defer func() {
			if err := recover(); err != nil {
				switch err.(type) {
				case string:
					if strings.Contains(err.(string), "code = Unauthenticated") {
						log.Printf("User is unauthenticated")
						http.Error(w, "User is unauthenticated", http.StatusUnauthorized)
					} else {
						log.Printf("Interal error occurred: %s", err.(string))
						http.Error(w, err.(string), http.StatusInternalServerError)
					}
				case error:
					st, ok := status.FromError(err.(error))
					if ok {
						if st.Code() == 16 {
							log.Printf("User is unauthenticated")
							http.Error(w, "User is unauthenticated", http.StatusUnauthorized)
						}
					} else {
						log.Printf("Interal error occurred: %s", err.(error).Error())
						http.Error(w, err.(error).Error(), http.StatusInternalServerError)
					}
				default:
					http.Error(w, err.(string), http.StatusInternalServerError)
				}
			}
			return
		}()
		a.CheckRbacPermissions(SyncAction, ClusterResource, "abc")
		next.ServeHTTP(w, r)
	})
}

func (a *argoAuthenticatorApi) CheckRbacPermissions(action RbacAction, resource RbacResource, subresource string) bool {
	closer, client, err := a.configuredClient.NewAccountClient()
	if err != nil {
		log.Fatalf("account client could not be made for Argo: %s", err)
	}
	defer func() {
		if err = closer.Close(); err != nil {
			log.Println("failed to close argocd client: ", err.Error())
		}
	}()

	canI, err := client.CanI(context.TODO(), &account.CanIRequest{
		Resource:    string(resource),
		Action:      string(action),
		Subresource: subresource,
	})
	if err != nil {
		log.Fatalf("Request to Argo server failed: %s", err)
	}
	return canI.Value == "yes"
}

func getClientCreationData() string {
	argoCdServer := os.Getenv(apiclient.EnvArgoCDServer)
	if argoCdServer == "" {
		argoCdServer = DefaultApiServerAddress
	}
	return argoCdServer
}

func (a *argoAuthenticatorApi) initArgoClient() {
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

func (a *argoAuthenticatorApi) initArgoTLSCert() (string, error) {
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

func (a *argoAuthenticatorApi) configureArgoClient(token string) {
	argoClient, err := apiclient.NewClient(a.getAPIClientOptions(token))
	if err != nil {
		log.Fatalf("Error when making properly authenticated client: %s", err)
	}
	a.configuredClient = argoClient
}

func (a *argoAuthenticatorApi) getAPIClientOptions(token string) *apiclient.ClientOptions {
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

func (a *argoAuthenticatorApi) watchArgoTLSUpdates() error {
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

func (a *argoAuthenticatorApi) updateArgoTLSCert(certPEM []byte) error {
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
