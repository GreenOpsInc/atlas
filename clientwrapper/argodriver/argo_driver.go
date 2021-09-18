package argodriver

import (
	"context"
	"fmt"
	"github.com/argoproj/argo-cd/pkg/apiclient"
	"github.com/argoproj/argo-cd/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/util/config"
	"github.com/argoproj/argo-cd/util/io"
	"github.com/argoproj/gitops-engine/pkg/health"
	"greenops.io/client/atlasoperator/requestdatatypes"
	"greenops.io/client/k8sdriver"
	"greenops.io/client/progressionchecker/datamodel"
	"greenops.io/client/util"
	utilpointer "k8s.io/utils/pointer"
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

const (
	ArgoRevisionHashLatest string = "LATEST_REVISION"
)

const (
	ArgoSyncStrategyDefault          string = "N/A"
	ArgoSyncStrategyForceDefault     bool   = false
	ArgoSyncPruneDefault             bool   = false
	ArgoSyncSelectiveDefault         bool   = false
	AnnotationAtlasSyncStrategy      string = "atlas-argo-sync-strategy"
	AnnotationAtlasSyncStrategyForce string = "atlas-argo-sync-strategy-force"
	AnnotationAtlasSyncPrune         string = "atlas-argo-sync-prune"
	AnnotationAtlasSyncSelective     string = "atlas-argo-sync-selective"
)

type ArgoGetRestrictedClient interface {
	GetAppResourcesStatus(applicationName string) ([]datamodel.ResourceStatus, error)
	GetOperationSuccess(applicationName string) (bool, bool, string, error)
	GetCurrentRevisionHash(applicationName string) (string, error)
	GetLatestRevision(applicationName string) (int64, error)
}

type ArgoClient interface {
	Deploy(configPayload *string, revisionHash string) (bool, string, string, string)
	Sync(applicationName string) (bool, string, string, string)
	SelectiveSync(applicationName string, revisionHash string, gvkGroup requestdatatypes.GvkGroupRequest) (bool, string, string, string)
	GetAppResourcesStatus(applicationName string) ([]datamodel.ResourceStatus, error)
	GetOperationSuccess(applicationName string) (bool, bool, string, error)
	GetCurrentRevisionHash(applicationName string) (string, error)
	GetLatestRevision(applicationName string) (int64, error)
	Delete(applicationName string) bool
	Rollback(appName string, appRevisionId string) (bool, string, string, string)
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

func (a ArgoClientDriver) Deploy(configPayload *string, revisionHash string) (bool, string, string, string) {
	ioCloser, applicationClient, err := a.client.NewApplicationClient()
	if err != nil {
		log.Printf("The deploy application client could not be made. Error was %s\n", err)
		return false, "", "", ""
	}
	defer ioCloser.Close()

	applicationPayload := makeApplication(configPayload)
	//This "locks" the Argo applications under a specific pipeline run. Updates to the Argo app
	//in the middle of a run would be quite messy.
	if revisionHash != ArgoRevisionHashLatest {
		applicationPayload.Spec.Source.TargetRevision = revisionHash
	}
	_, err = a.kubernetesClient.CheckAndCreateNamespace(applicationPayload.Spec.Destination.Namespace)
	if err != nil {
		return false, "", "", ""
	}

	existingApp, err := applicationClient.Get(context.TODO(), &application.ApplicationQuery{Name: utilpointer.StringPtr(applicationPayload.Name)})
	if err == nil {
		log.Printf("Argo application named %s already exists\n", existingApp.Name)
		if a.SpecMatches(&applicationPayload, existingApp) {
			log.Printf("Specs match, triggering sync...\n")
			return a.Sync(existingApp.Name)
		} else {
			log.Printf("Specs differ, triggering update...\n")
			return a.Update(configPayload)
		}
	}

	argoApplication, err := applicationClient.Create(
		context.TODO(),
		&application.ApplicationCreateRequest{Application: applicationPayload},
	) //CallOption is not necessary, for now...
	if err != nil {
		log.Printf("The deploy step threw an error. Error was %s\n", err)
		return false, "", "", ""
	}

	log.Printf("Deploying Argo application named %s\n", argoApplication.Name)
	return a.Sync(argoApplication.Name)
}

func (a ArgoClientDriver) Update(configPayload *string) (bool, string, string, string) {
	ioCloser, applicationClient, err := a.client.NewApplicationClient()
	if err != nil {
		log.Printf("The deploy application client could not be made. Error was %s\n", err)
		return false, "", "", ""
	}
	defer ioCloser.Close()

	applicationPayload := makeApplication(configPayload)
	_, err = a.kubernetesClient.CheckAndCreateNamespace(applicationPayload.Spec.Destination.Namespace)
	if err != nil {
		return false, "", "", ""
	}

	argoApplication, err := applicationClient.Update(
		context.TODO(),
		&application.ApplicationUpdateRequest{
			Application: &applicationPayload,
		},
	) //CallOption is not necessary, for now...
	if err != nil {
		log.Printf("The update step threw an error. Error was %s\n", err)
		return false, "", "", ""
	}

	log.Printf("Updated & now syncing Argo application named %s\n", argoApplication.Name)
	return a.Sync(argoApplication.Name)
}

func (a ArgoClientDriver) updateApplicationSourceHash(applicationName string, revisionHash string) (bool, string, string, string) {
	ioCloser, applicationClient, err := a.client.NewApplicationClient()
	if err != nil {
		log.Printf("The deploy application client could not be made. Error was %s\n", err)
		return false, "", "", ""
	}
	defer ioCloser.Close()

	applicationPayload, err := applicationClient.Get(context.TODO(), &application.ApplicationQuery{Name: &applicationName})
	if err != nil {
		return false, "", "", ""
	}
	_, err = a.kubernetesClient.CheckAndCreateNamespace(applicationPayload.Spec.Destination.Namespace)
	if err != nil {
		return false, "", "", ""
	}

	applicationPayload.Spec.Source.TargetRevision = revisionHash
	argoApplication, err := applicationClient.Update(
		context.TODO(),
		&application.ApplicationUpdateRequest{
			Application: applicationPayload,
		},
	) //CallOption is not necessary, for now...
	if err != nil {
		log.Printf("The update step threw an error. Error was %s\n", err)
		return false, "", "", ""
	}

	log.Printf("Updated & now syncing Argo application named %s\n", argoApplication.Name)
	return a.Sync(argoApplication.Name)
}

func (a ArgoClientDriver) Sync(applicationName string) (bool, string, string, string) {
	ioCloser, applicationClient, err := a.client.NewApplicationClient()
	if err != nil {
		log.Printf("The deploy application client could not be made. Error was %s\n", err)
		return false, "", "", ""
	}
	defer ioCloser.Close()
	//Sync() returns the current state of the application and triggers the synchronization of the application, so the return
	//value is not useful in this case
	var argoApplication *v1alpha1.Application
	argoApplication, err = applicationClient.Get(context.TODO(), &application.ApplicationQuery{Name: &applicationName})
	if err != nil {
		log.Printf("Getting the application threw an error. Error was %s\n", err)
		return false, "", "", ""
	}
	prune := ArgoSyncPruneDefault
	strategy := ArgoSyncStrategyDefault
	selectiveSync := ArgoSyncSelectiveDefault
	syncStrategyForce := ArgoSyncStrategyForceDefault
	if val, ok := argoApplication.Annotations[AnnotationAtlasSyncPrune]; ok {
		if val == "true" {
			prune = true
		}
	}
	if val, ok := argoApplication.Annotations[AnnotationAtlasSyncStrategy]; ok {
		strategy = val
	}
	if val, ok := argoApplication.Annotations[AnnotationAtlasSyncSelective]; ok {
		if val == "true" {
			selectiveSync = true
		}
	}
	if val, ok := argoApplication.Annotations[AnnotationAtlasSyncStrategyForce]; ok {
		if val == "true" {
			syncStrategyForce = true
		}
	}
	var applicationSyncRequest application.ApplicationSyncRequest
	applicationSyncRequest = application.ApplicationSyncRequest{
		Name:          &applicationName,
		Prune:         prune,
	}
	if argoApplication.Spec.SyncPolicy != nil && argoApplication.Spec.SyncPolicy.Retry != nil {
		applicationSyncRequest.RetryStrategy = argoApplication.Spec.SyncPolicy.Retry
	}
	if selectiveSync {
		resourceStatuses := make([]v1alpha1.SyncOperationResource, 0)
		for _, resource := range argoApplication.Status.Resources {
			requiresSyncingHealthList := []health.HealthStatusCode{
				health.HealthStatusMissing, health.HealthStatusUnknown, health.HealthStatusDegraded, health.HealthStatusSuspended,
			}
			requiresSyncingSyncList := []v1alpha1.SyncStatusCode{
				v1alpha1.SyncStatusCodeUnknown, v1alpha1.SyncStatusCodeOutOfSync,
			}
			if containsHealth(requiresSyncingHealthList, resource.Health.Status) || containsSync(requiresSyncingSyncList, resource.Status) {
				resourceStatuses = append(resourceStatuses, v1alpha1.SyncOperationResource{
					Group:     resource.Group,
					Kind:      resource.Kind,
					Name:      resource.Name,
					Namespace: resource.Namespace,
				})
			}
		}
		applicationSyncRequest.Resources = resourceStatuses
	}
	switch strategy {
	case "apply":
		applicationSyncRequest.Strategy = &v1alpha1.SyncStrategy{Apply: &v1alpha1.SyncStrategyApply{}}
		applicationSyncRequest.Strategy.Apply.Force = syncStrategyForce
	case "", "hook":
		applicationSyncRequest.Strategy = &v1alpha1.SyncStrategy{Hook: &v1alpha1.SyncStrategyHook{}}
		applicationSyncRequest.Strategy.Hook.Force = syncStrategyForce
	default:
		log.Printf("Unknown sync strategy: '%s'. Using default...", strategy)
	}
	argoApplication, err = applicationClient.Sync(context.TODO(), &applicationSyncRequest)
	if err != nil {
		log.Printf("Syncing threw an error. Error was %s\n", err)
		return false, "", "", ""
	}
	log.Printf("Triggered sync for Argo application named %s\n", argoApplication.Name)
	return true, argoApplication.Name, argoApplication.Namespace, argoApplication.Operation.Sync.Revision
}

func (a ArgoClientDriver) SelectiveSync(applicationName string, revisionHash string, gvkGroup requestdatatypes.GvkGroupRequest) (bool, string, string, string) {
	ioCloser, applicationClient, err := a.client.NewApplicationClient()
	if err != nil {
		log.Printf("The deploy application client could not be made. Error was %s\n", err)
		return false, "", "", ""
	}
	defer ioCloser.Close()

	var argoApplication *v1alpha1.Application
	argoApplication, err = applicationClient.Get(context.TODO(), &application.ApplicationQuery{Name: &applicationName})
	if err != nil {
		log.Printf("Getting the application threw an error. Error was %s\n", err)
		return false, "", "", ""
	}
	prune := ArgoSyncPruneDefault
	if val, ok := argoApplication.Annotations[AnnotationAtlasSyncPrune]; ok {
		if val == "true" {
			prune = true
		}
	}
	resources := make([]string, 0)
	for _, value := range gvkGroup.ResourceList {
		resources = append(resources, fmt.Sprintf("%s:%s:%s:%s", value.Group, value.Kind, value.ResourceName, value.ResourceNamespace))
	}
	selectedResources, err := a.ParseSelectedResources(resources)
	if err != nil {
		log.Printf("Resources could not be parsed. Error was %s\n", err)
		return false, "", "", ""
	}
	syncReq := application.ApplicationSyncRequest{
		Name:     &applicationName,
		Revision: revisionHash,
		Prune:    prune,
	}
	if len(selectedResources) > 0 {
		syncReq.Resources = selectedResources
	}
	argoApplication, err = applicationClient.Sync(context.TODO(), &syncReq)
	if err != nil {
		log.Printf("Syncing threw an error. Error was %s\n", err)
		return false, "", "", ""
	}
	log.Printf("Triggered selective sync for Argo application named %s\n", argoApplication.Name)
	return true, argoApplication.Name, argoApplication.Namespace, revisionHash
}

func (a ArgoClientDriver) GetAppResourcesStatus(applicationName string) ([]datamodel.ResourceStatus, error) {
	ioCloser, applicationClient, err := a.client.NewApplicationClient()
	resourceStatuses := make([]datamodel.ResourceStatus, 0)
	if err != nil {
		log.Printf("The deploy application client could not be made. Error was %s\n", err)
		return resourceStatuses, err
	}
	defer ioCloser.Close()
	app, err := applicationClient.Get(context.TODO(), &application.ApplicationQuery{Name: &applicationName})
	if err != nil {
		if strings.Contains(err.Error(), "\""+applicationName+"\" not found") {
			//TODO: Probably need better handling for this
			return resourceStatuses, err
		}
		log.Printf("Getting the application threw an error. Error was %s\n", err)
		return resourceStatuses, err
	}

	for _, resource := range app.Status.Resources {
		if resource.Health.Status != health.HealthStatusHealthy || resource.Status != v1alpha1.SyncStatusCodeSynced {
			resourceStatuses = append(resourceStatuses, datamodel.ResourceStatus{
				GroupVersionKind:  resource.GroupVersionKind(),
				ResourceName:      resource.Name,
				ResourceNamespace: resource.Namespace,
				HealthStatus:      string(resource.Health.Status),
				SyncStatus:        string(resource.Status),
			})
		}
	}

	return resourceStatuses, nil
}

func (a ArgoClientDriver) GetOperationSuccess(applicationName string) (bool, bool, string, error) {
	ioCloser, applicationClient, err := a.client.NewApplicationClient()
	if err != nil {
		log.Printf("The deploy application client could not be made. Error was %s\n", err)
		return false, false, "", err
	}
	defer ioCloser.Close()
	app, err := applicationClient.Get(context.TODO(), &application.ApplicationQuery{Name: &applicationName})
	if err != nil {
		if strings.Contains(err.Error(), "\""+applicationName+"\" not found") {
			//TODO: Probably need better handling for this
			return false, false, "", nil
		}
		log.Printf("Getting the application threw an error. Error was %s\n", err)
		return false, false, "", err
	}

	if app.Status.OperationState == nil || app.Status.OperationState.SyncResult == nil {
		return false, false, "", nil
	}
	return app.Status.OperationState.Phase.Completed(),
		app.Status.OperationState.Phase.Successful(),
		app.Status.OperationState.SyncResult.Revision,
		nil
}

func (a ArgoClientDriver) GetCurrentRevisionHash(applicationName string) (string, error) {
	ioCloser, applicationClient, err := a.client.NewApplicationClient()
	if err != nil {
		log.Printf("The deploy application client could not be made. Error was %s\n", err)
		return "", err
	}
	defer ioCloser.Close()
	app, err := applicationClient.Get(context.TODO(), &application.ApplicationQuery{Name: &applicationName})
	if err != nil {
		if strings.Contains(err.Error(), "\""+applicationName+"\" not found") {
			return "", nil
		}
		log.Printf("Getting the application threw an error. Error was %s\n", err)
		return "", err
	}
	//Note, this is the most recent SYNCED revision. If an operation is still progressing, the revision tied to that operation
	//will likely not be sent
	return app.Status.Sync.Revision, nil
}

func (a ArgoClientDriver) GetLatestRevision(applicationName string) (int64, error) {
	ioCloser, applicationClient, err := a.client.NewApplicationClient()
	if err != nil {
		log.Printf("The deploy application client could not be made. Error was %s\n", err)
		return -1, err
	}
	defer ioCloser.Close()
	app, err := applicationClient.Get(context.TODO(), &application.ApplicationQuery{Name: &applicationName})
	if err != nil {
		if strings.Contains(err.Error(), "\""+applicationName+"\" not found") {
			return -1, nil
		}
		log.Printf("Getting the application threw an error. Error was %s\n", err)
		return -1, err
	}
	return app.Status.History.LastRevisionHistory().ID, nil
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

func (a ArgoClientDriver) Rollback(appName string, appRevisionHash string) (bool, string, string, string) {
	ioCloser, applicationClient, err := a.client.NewApplicationClient()
	if err != nil {
		log.Printf("The rollback application client could not be made. Error was %s\n", err)
		return false, "", "", ""
	}
	defer ioCloser.Close()

	existingApp, err := applicationClient.Get(context.TODO(), &application.ApplicationQuery{Name: utilpointer.StringPtr(appName)})
	if err != nil {
		log.Printf("The rollback threw an error. Error was %s\n", err)
		return false, "", "", ""
	}

	deploymentHistory := existingApp.Status.History
	idx := len(deploymentHistory) - 1
	revisionId := int64(-1)
	for idx >= 0 {
		if deploymentHistory[idx].Revision == appRevisionHash {
			revisionId = deploymentHistory[idx].ID
			break
		}
		idx--
	}
	if revisionId == -1 {
		log.Printf("Couldn't find the expected hash in the Argo history...triggering an update with the relevant hash")
		return a.updateApplicationSourceHash(appName, appRevisionHash)
	}

	argoApplication, err := applicationClient.Rollback(
		context.TODO(),
		&application.ApplicationRollbackRequest{
			Name:  &appName,
			ID:    revisionId,
			Prune: true,
		},
	)
	if err != nil {
		log.Printf("The rollback threw an error. Error was %s\n", err)
		return false, "", "", ""
	}

	log.Printf("Rolled back Argo application named %s\n", argoApplication.Name)
	//TODO: Syncing takes time. Right now, we can assume that apps will deploy properly. In the future, we will have to see whether we can blindly return true or not.
	return true, argoApplication.Name, argoApplication.Namespace, argoApplication.Operation.Sync.Revision
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

func makeApplication(configPayload *string) v1alpha1.Application {
	var argoApplication v1alpha1.Application
	err := config.UnmarshalReader(strings.NewReader(*configPayload), &argoApplication)
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
