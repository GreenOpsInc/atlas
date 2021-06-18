package argodriver

import (
	"context"
	"github.com/argoproj/argo-cd/pkg/apiclient"
	"github.com/argoproj/argo-cd/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/util/config"
	"github.com/argoproj/argo-cd/util/io"
	"github.com/argoproj/gitops-engine/pkg/health"
	"greenops.io/client/util"
	"log"
	"os"
	"strings"

	sessionpkg "github.com/argoproj/argo-cd/pkg/apiclient/session"
	grpcutil "github.com/argoproj/argo-cd/util/grpc"
)

const (
	UserAccountEnvVar         string = "ARGOCD_USER_ACCOUNT"
	UserAccountPasswordEnvVar string = "ARGOCD_USER_PASSWORD"
	DefaultApiServerAddress   string = "argocd-server.argocd.svc.cluster.local"
	DefaultUserAccount        string = "admin"
)

type ArgoClient interface {
	//TODO: Add parameters for Deploy
	Deploy(configPayload string) bool
	//TODO: Add parameters for Delete
	Delete(configPayload string) bool
	//TODO: Update parameters & return type for CheckStatus
	CheckStatus() bool
}

type ArgoClientDriver struct {
	client apiclient.Client
}

//TODO: ALL functions should have a callee tag on them
func New() ArgoClient {
	apiServerAddress, userAccount, userPassword, _ := getClientCreationData()
	tlsTestResult, err := grpcutil.TestTLS(apiServerAddress)
	util.CheckFatalError(err)

	argoClient, err := apiclient.NewClient(
		&apiclient.ClientOptions{
			ServerAddr: apiServerAddress,
			Insecure:   true,
			PlainText:  !tlsTestResult.TLS,
		})
	util.CheckFatalError(err)

	closer, sessionClient, err := argoClient.NewSessionClient()
	util.CheckFatalError(err)
	defer io.Close(closer)

	sessionResponse, err := sessionClient.Create(context.TODO(), &sessionpkg.SessionCreateRequest{Username: userAccount, Password: userPassword})
	util.CheckFatalError(err)

	argoClient, err = apiclient.NewClient(&apiclient.ClientOptions{
		Insecure:   true,
		ServerAddr: apiServerAddress,
		AuthToken:  sessionResponse.Token,
		PlainText:  !tlsTestResult.TLS,
	})
	util.CheckFatalError(err)

	var client ArgoClient
	client = ArgoClientDriver{argoClient}
	return client
}

func (a ArgoClientDriver) Deploy(configPayload string) bool {
	ioCloser, applicationClient, err := a.client.NewApplicationClient()
	if err != nil {
		log.Printf("The deploy application client could not be made. Error was %s\n", err)
		return false
	}
	defer ioCloser.Close()

	argoApplication, err := applicationClient.Create(
		context.TODO(),
		&application.ApplicationCreateRequest{Application: makeApplication(configPayload)},
	) //CallOption is not necessary, for now...
	if err != nil {
		log.Printf("The deploy step threw an error. Error was %s\n", err)
		return false
	}

	argoApplication, err = applicationClient.Sync(context.TODO(), &application.ApplicationSyncRequest{Name: &argoApplication.Name})
	if err != nil {
		log.Printf("Syncing threw an error. Error was %s\n", err)
		return false
	}
	return argoApplication.Status.Health.Status == health.HealthStatusHealthy && argoApplication.Status.Sync.Status == v1alpha1.SyncStatusCodeSynced
}

func (a ArgoClientDriver) Delete(configPayload string) bool {
	panic("implement me")
}

func (a ArgoClientDriver) CheckStatus() bool {
	panic("implement me")
}

func makeApplication(configPayload string) v1alpha1.Application {
	var argoApplication v1alpha1.Application
	err := config.UnmarshalReader(strings.NewReader(configPayload), &argoApplication)
	if err != nil {
		log.Printf("The unmarshalling step threw an error. Error was %s\n", err)
		return v1alpha1.Application{}
	}
	return argoApplication
}

func getClientCreationData() (string, string, string, string) {
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
		panic("The password has to be entered. The default is stored in a secret called argocd-initial-admin-secret " +
			"and can be fetched using 'kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath=\"{.data.password}\" | base64 -d'. " +
			"Fetching the password is not implemented yet.")
	}
	argoCdUserToken := os.Getenv(apiclient.EnvArgoCDAuthToken)
	if argoCdUserToken == "" {
		panic("An acces token has to be entered. Not implemented yet.")
	}
	return argoCdServer, argoCdUser, argoCdUserPassword, argoCdUserToken
}
