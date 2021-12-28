package argoauthenticator

import (
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

type ArgoAuthenticatorApi interface {
	CheckRbacPermissions(action RbacAction, resource RbacResource, subresource string) bool
}

type ArgoAuthenticatorApiImpl struct {
	apiServerAddress string
	tls              bool
	rawClient        apiclient.Client
	configuredClient apiclient.Client
}

func New() ArgoAuthenticatorApi {
	apiServerAddress := getClientCreationData()
	argoClient, tls := getRawArgoClient(apiServerAddress)

	var client ArgoAuthenticatorApi
	client = &ArgoAuthenticatorApiImpl{apiServerAddress: apiServerAddress, tls: tls, rawClient: argoClient, configuredClient: nil}
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

func (a *ArgoAuthenticatorApiImpl) getConfiguredArgoClient(token string) {
	log.Println("in getConfiguredArgoClient")
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

func (a *ArgoAuthenticatorApiImpl) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("in middleware http.HandlerFunc")
		token := r.Header.Get("Authorization")
		splitToken := strings.Split(token, "Bearer ")
		token = splitToken[1]
		a.getConfiguredArgoClient(token)

		defer func() {
			if err := recover(); err != nil {
				log.Println("in recover err: ", err)
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
						log.Println("in the error clause: ", err)
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

func (a *ArgoAuthenticatorApiImpl) CheckRbacPermissions(action RbacAction, resource RbacResource, subresource string) bool {
	log.Println("in CheckRbacPermissions")
	closer, client, err := a.configuredClient.NewAccountClient()
	if err != nil {
		panic(fmt.Sprintf("account client could not be made for Argo: %s", err))
	}
	defer closer.Close()
	canI, err := client.CanI(context.TODO(), &account.CanIRequest{
		Resource:    string(resource),
		Action:      string(action),
		Subresource: subresource,
	})
	if err != nil {
		panic(fmt.Sprintf("Request to Argo server failed: %s", err))
	}
	return canI.Value == "yes"
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
