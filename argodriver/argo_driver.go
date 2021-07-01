package argodriver

import (
	"context"
	"github.com/argoproj/argo-cd/pkg/apiclient"
	"github.com/argoproj/argo-cd/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/util/config"
	"github.com/argoproj/argo-cd/util/io"
	"github.com/argoproj/gitops-engine/pkg/health"
	"greenops.io/client/k8sdriver"
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
	Deploy(configPayload string) (bool, string)
	//TODO: Add parameters for Delete
	Delete(applicationName string) bool
	//TODO: Update parameters & return type for CheckStatus
	CheckHealthy(argoApplicationName string) bool
}

type ArgoClientDriver struct {
	client           apiclient.Client
	kubernetesClient k8sdriver.KubernetesClientNamespaceRestricted
}

//TODO: ALL functions should have a callee tag on them
func New(kubernetesDriver *k8sdriver.KubernetesClient) ArgoClient {
	apiServerAddress, userAccount, userPassword, _ := getClientCreationData(kubernetesDriver)
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
	var kubernetesClient k8sdriver.KubernetesClientNamespaceRestricted
	kubernetesClient = *kubernetesDriver
	client = ArgoClientDriver{argoClient, kubernetesClient}
	return client
}

func (a ArgoClientDriver) Deploy(configPayload string) (bool, string) {
	ioCloser, applicationClient, err := a.client.NewApplicationClient()
	if err != nil {
		log.Printf("The deploy application client could not be made. Error was %s\n", err)
		return false, ""
	}
	defer ioCloser.Close()

	applicationPayload := makeApplication(configPayload)
	err = a.kubernetesClient.CheckAndCreateNamespace(applicationPayload.Spec.Destination.Namespace)
	if err != nil {
		return false, ""
	}

	argoApplication, err := applicationClient.Create(
		context.TODO(),
		&application.ApplicationCreateRequest{Application: applicationPayload},
	) //CallOption is not necessary, for now...
	if err != nil {
		log.Printf("The deploy step threw an error. Error was %s\n", err)
		return false, ""
	}

	//Sync() returns the current state of the application and triggers the synchronization of the application, so the return
	//value is not useful in this case
	_, err = applicationClient.Sync(context.TODO(), &application.ApplicationSyncRequest{Name: &argoApplication.Name})
	if err != nil {
		log.Printf("Syncing threw an error. Error was %s\n", err)
		return false, ""
	}
	log.Printf("Deployed Argo application named %s\n", argoApplication.Name)
	//TODO: Syncing takes time. Right now, we can assume that apps will deploy properly. In the future, we will have to see whether we can blindly return true or not.
	return true, argoApplication.Namespace
}

func (a ArgoClientDriver) Delete(applicationName string) bool {
	ioCloser, applicationClient, err := a.client.NewApplicationClient()
	if err != nil {
		log.Printf("The deploy application client could not be made. Error was %s\n", err)
		return false
	}
	defer ioCloser.Close()
	_, err = applicationClient.Delete(context.TODO(), &application.ApplicationDeleteRequest{Name: &applicationName})
	if err != nil && !strings.Contains(err.Error(), "\""+applicationName+"\" not found") {
		log.Printf("Deletion threw an error. Error was %s\n", err)
		return false
	}
	//TODO: Deleting takes time. Right now, we can assume that apps will delete properly. In the future, we will have to see whether we can blindly return true or not.
	return true
}

func (a ArgoClientDriver) CheckHealthy(argoApplicationName string) bool {
	ioCloser, applicationClient, err := a.client.NewApplicationClient()
	if err != nil {
		log.Printf("The application client could not be made. Error was %s\n", err)
		return false
	}
	defer ioCloser.Close()
	//Sync() returns the current state of the application and triggers the synchronization of the application, so the return
	//value is the state BEFORE the sync
	argoApplication, err := applicationClient.Sync(context.TODO(), &application.ApplicationSyncRequest{Name: &argoApplicationName})
	if err != nil {
		log.Printf("Syncing threw an error. Error was %s\n", err)
		return false
	}
	return argoApplication.Status.Health.Status == health.HealthStatusHealthy && argoApplication.Status.Sync.Status == v1alpha1.SyncStatusCodeSynced
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

func getClientCreationData(kubernetesDriver *k8sdriver.KubernetesClient) (string, string, string, string) {
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
