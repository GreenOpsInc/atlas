package argodriver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/argoproj/argo-cd/pkg/apiclient"
	"github.com/argoproj/argo-cd/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/pkg/apiclient/project"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/util/config"
	"github.com/argoproj/gitops-engine/pkg/health"
	"github.com/dgrijalva/jwt-go/v4"
	"github.com/greenopsinc/util/clientrequest"
	"github.com/greenopsinc/util/tlsmanager"
	"greenops.io/client/k8sdriver"
	"greenops.io/client/progressionchecker/datamodel"
	"greenops.io/client/util"
	utilpointer "k8s.io/utils/pointer"
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
	Deploy(configPayload *string, revisionHash string) (string, string, string, error)
	Sync(applicationName string) (string, string, string, error)
	SelectiveSync(applicationName string, revisionHash string, gvkGroup clientrequest.GvkGroupRequest) (string, string, string, error)
	GetAppResourcesStatus(applicationName string) ([]datamodel.ResourceStatus, error)
	GetOperationSuccess(applicationName string) (bool, bool, string, error)
	GetCurrentRevisionHash(applicationName string) (string, error)
	GetLatestRevision(applicationName string) (int64, error)
	Delete(applicationName string) error
	Rollback(appName string, appRevisionId string) (string, string, string, error)
	MarkNoDeploy(cluster string, namespace string, apply bool) error
	//TODO: Update parameters & return type for CheckStatus
	CheckHealthy(argoApplicationName string) bool
}

type ArgoAuthClient interface {
	GetAuthToken() string
}

type ArgoClientDriver struct {
	client           apiclient.Client
	tm               tlsmanager.Manager
	tlsCertPath      string
	tlsEnabled       bool
	apiServerAddress string
	kubernetesClient k8sdriver.KubernetesClientNamespaceSecretRestricted
}

//TODO: ALL functions should have a callee tag on them
func New(kubernetesDriver *k8sdriver.KubernetesClient, tm tlsmanager.Manager) ArgoClient {
	var kubernetesClient k8sdriver.KubernetesClientNamespaceSecretRestricted
	kubernetesClient = *kubernetesDriver
	apiServerAddress, userAccount, userPassword, _ := getClientCreationData(&kubernetesClient)
	driver := &ArgoClientDriver{kubernetesClient: kubernetesClient, apiServerAddress: apiServerAddress, tm: tm}
	if err := driver.initArgoDriver(userAccount, userPassword); err != nil {
		util.CheckFatalError(err)
	}
	return driver
}

func (a *ArgoClientDriver) CheckForRefresh() error {
	parser := &jwt.Parser{
		ValidationHelper: jwt.NewValidationHelper(jwt.WithoutClaimsValidation(), jwt.WithoutAudienceValidation()),
	}
	var claims jwt.StandardClaims
	_, _, err := parser.ParseUnverified(a.client.ClientOptions().AuthToken, &claims)
	if err != nil {
		log.Printf("Getting token claims failed with error: %s", err)
		return err
	}
	now := jwt.At(time.Now().UTC())
	if !now.Time.After(claims.ExpiresAt.Time) {
		return nil
	}

	log.Printf("Getting new token")
	_, userAccount, userPassword, _ := getClientCreationData(&a.kubernetesClient)
	token, err := a.generateArgoToken(userAccount, userPassword)
	if err != nil {
		return err
	}
	argoClient, err := apiclient.NewClient(a.getAPIClientOptions(token))
	if err != nil {
		return errors.New(fmt.Sprintf("error when making properly authenticated client: %s", err.Error()))
	}
	a.client = argoClient
	return nil
}

func (a *ArgoClientDriver) GetAuthToken() string {
	err := a.CheckForRefresh()
	if err != nil {
		log.Printf("Error when getting auth token, returning empty string")
		return ""
	}
	return a.client.ClientOptions().AuthToken
}

func (a *ArgoClientDriver) Deploy(configPayload *string, revisionHash string) (string, string, string, error) {
	err := a.CheckForRefresh()
	if err != nil {
		return "", "", "", err
	}
	ioCloser, applicationClient, err := a.client.NewApplicationClient()
	if err != nil {
		log.Printf("The deploy application client could not be made. Error was %s\n", err)
		return "", "", "", err
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
		return "", "", "", err
	}

	existingApp, err := applicationClient.Get(context.TODO(), &application.ApplicationQuery{Name: utilpointer.StringPtr(applicationPayload.Name)})
	if err == nil {
		log.Printf("Argo application named %s already exists\n", existingApp.Name)
		if a.SpecMatches(&applicationPayload, existingApp) {
			log.Printf("Specs match, triggering sync...\n")
			return a.Sync(existingApp.Name)
		} else {
			log.Printf("Specs differ, triggering update...\n")
			return a.Update(&applicationPayload)
		}
	}

	argoApplication, err := applicationClient.Create(
		context.TODO(),
		&application.ApplicationCreateRequest{Application: applicationPayload},
	) //CallOption is not necessary, for now...
	if err != nil {
		log.Printf("The deploy step threw an error. Error was %s\n", err)
		return "", "", "", err
	}

	log.Printf("Deploying Argo application named %s\n", argoApplication.Name)
	return a.Sync(argoApplication.Name)
}

func (a *ArgoClientDriver) Update(applicationPayload *v1alpha1.Application) (string, string, string, error) {
	err := a.CheckForRefresh()
	if err != nil {
		return "", "", "", err
	}
	ioCloser, applicationClient, err := a.client.NewApplicationClient()
	if err != nil {
		log.Printf("The deploy application client could not be made. Error was %s\n", err)
		return "", "", "", err
	}
	defer ioCloser.Close()

	_, err = a.kubernetesClient.CheckAndCreateNamespace(applicationPayload.Spec.Destination.Namespace)
	if err != nil {
		return "", "", "", err
	}

	argoApplication, err := applicationClient.Update(
		context.TODO(),
		&application.ApplicationUpdateRequest{
			Application: applicationPayload,
		},
	) //CallOption is not necessary, for now...
	if err != nil {
		log.Printf("The update step threw an error. Error was %s\n", err)
		return "", "", "", err
	}

	log.Printf("Updated & now syncing Argo application named %s\n", argoApplication.Name)
	return a.Sync(argoApplication.Name)
}

func (a *ArgoClientDriver) updateApplicationSourceHash(applicationName string, revisionHash string) (string, string, string, error) {
	err := a.CheckForRefresh()
	if err != nil {
		return "", "", "", err
	}
	ioCloser, applicationClient, err := a.client.NewApplicationClient()
	if err != nil {
		log.Printf("The deploy application client could not be made. Error was %s\n", err)
		return "", "", "", err
	}
	defer ioCloser.Close()

	applicationPayload, err := applicationClient.Get(context.TODO(), &application.ApplicationQuery{Name: &applicationName})
	if err != nil {
		return "", "", "", err
	}
	_, err = a.kubernetesClient.CheckAndCreateNamespace(applicationPayload.Spec.Destination.Namespace)
	if err != nil {
		return "", "", "", err
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
		return "", "", "", err
	}

	log.Printf("Updated & now syncing Argo application named %s\n", argoApplication.Name)
	return a.Sync(argoApplication.Name)
}

func (a *ArgoClientDriver) Sync(applicationName string) (string, string, string, error) {
	err := a.CheckForRefresh()
	if err != nil {
		return "", "", "", err
	}
	ioCloser, applicationClient, err := a.client.NewApplicationClient()
	if err != nil {
		log.Printf("The deploy application client could not be made. Error was %s\n", err)
		return "", "", "", err
	}
	defer ioCloser.Close()
	//Sync() returns the current state of the application and triggers the synchronization of the application, so the return
	//value is not useful in this case
	var argoApplication *v1alpha1.Application
	argoApplication, err = applicationClient.Get(context.TODO(), &application.ApplicationQuery{Name: &applicationName})
	if err != nil {
		log.Printf("Getting the application threw an error. Error was %s\n", err)
		return "", "", "", err
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
		Name:  &applicationName,
		Prune: prune,
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
		return "", "", "", err
	}
	log.Printf("Triggered sync for Argo application named %s\n", argoApplication.Name)
	return argoApplication.Name, argoApplication.Namespace, argoApplication.Operation.Sync.Revision, nil
}

func (a *ArgoClientDriver) SelectiveSync(applicationName string, revisionHash string, gvkGroup clientrequest.GvkGroupRequest) (string, string, string, error) {
	err := a.CheckForRefresh()
	if err != nil {
		return "", "", "", err
	}
	ioCloser, applicationClient, err := a.client.NewApplicationClient()
	if err != nil {
		log.Printf("The deploy application client could not be made. Error was %s\n", err)
		return "", "", "", err
	}
	defer ioCloser.Close()

	var argoApplication *v1alpha1.Application
	argoApplication, err = applicationClient.Get(context.TODO(), &application.ApplicationQuery{Name: &applicationName})
	if err != nil {
		log.Printf("Getting the application threw an error. Error was %s\n", err)
		return "", "", "", err
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
		return "", "", "", err
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
		return "", "", "", err
	}
	log.Printf("Triggered selective sync for Argo application named %s\n", argoApplication.Name)
	return argoApplication.Name, argoApplication.Namespace, revisionHash, nil
}

func (a *ArgoClientDriver) GetAppResourcesStatus(applicationName string) ([]datamodel.ResourceStatus, error) {
	err := a.CheckForRefresh()
	if err != nil {
		return nil, err
	}
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

func (a *ArgoClientDriver) GetOperationSuccess(applicationName string) (bool, bool, string, error) {
	err := a.CheckForRefresh()
	if err != nil {
		return false, false, "", err
	}
	ioCloser, applicationClient, err := a.client.NewApplicationClient()
	if err != nil {
		log.Printf("The deploy application client could not be made. Error was %s\n", err)
		return false, false, "", err
	}
	defer ioCloser.Close()
	app, err := applicationClient.Get(context.TODO(), &application.ApplicationQuery{Name: &applicationName})
	if err != nil {
		log.Printf("Getting the application threw an error. Error was %s\n", err)
		return false, false, "", err
	}

	if app.Status.OperationState == nil || app.Status.OperationState.SyncResult == nil {
		return false, false, "", errors.New("couldn't get status information")
	}
	return app.Status.OperationState.Phase.Completed(),
		app.Status.OperationState.Phase.Successful(),
		app.Status.OperationState.SyncResult.Revision,
		nil
}

func (a *ArgoClientDriver) GetCurrentRevisionHash(applicationName string) (string, error) {
	err := a.CheckForRefresh()
	if err != nil {
		return "", err
	}
	ioCloser, applicationClient, err := a.client.NewApplicationClient()
	if err != nil {
		log.Printf("The deploy application client could not be made. Error was %s\n", err)
		return "", err
	}
	defer ioCloser.Close()
	app, err := applicationClient.Get(context.TODO(), &application.ApplicationQuery{Name: &applicationName})
	if err != nil {
		log.Printf("Getting the application threw an error. Error was %s\n", err)
		return "", err
	}
	//Note, this is the most recent SYNCED revision. If an operation is still progressing, the revision tied to that operation
	//will likely not be sent
	if app.Status.Sync.Revision == "" {
		return "", errors.New("argo revision is empty")
	}
	return app.Status.Sync.Revision, nil
}

func (a *ArgoClientDriver) GetLatestRevision(applicationName string) (int64, error) {
	err := a.CheckForRefresh()
	if err != nil {
		return 0, err
	}
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

func (a *ArgoClientDriver) Delete(applicationName string) error {
	err := a.CheckForRefresh()
	if err != nil {
		return err
	}
	ioCloser, applicationClient, err := a.client.NewApplicationClient()
	if err != nil {
		log.Printf("The deploy application client could not be made. Error was %s\n", err)
		return err
	}
	defer ioCloser.Close()
	_, err = applicationClient.Delete(context.TODO(), &application.ApplicationDeleteRequest{Name: &applicationName})
	if err != nil && !strings.Contains(err.Error(), "\""+applicationName+"\" not found") {
		log.Printf("Deletion threw an error. Error was %s\n", err)
		return err
	}
	//TODO: Deleting takes time. Right now, we can assume that apps will delete properly. In the future, we will have to see whether we can blindly return true or not.
	return nil
}

func (a *ArgoClientDriver) Rollback(appName string, appRevisionHash string) (string, string, string, error) {
	err := a.CheckForRefresh()
	if err != nil {
		return "", "", "", err
	}
	ioCloser, applicationClient, err := a.client.NewApplicationClient()
	if err != nil {
		log.Printf("The rollback application client could not be made. Error was %s\n", err)
		return "", "", "", err
	}
	defer ioCloser.Close()

	existingApp, err := applicationClient.Get(context.TODO(), &application.ApplicationQuery{Name: utilpointer.StringPtr(appName)})
	if err != nil {
		log.Printf("The rollback threw an error. Error was %s\n", err)
		return "", "", "", err
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
		return "", "", "", err
	}

	log.Printf("Rolled back Argo application named %s\n", argoApplication.Name)
	//TODO: Syncing takes time. Right now, we can assume that apps will deploy properly. In the future, we will have to see whether we can blindly return true or not.
	return argoApplication.Name, argoApplication.Namespace, argoApplication.Operation.Sync.Revision, nil
}

func (a *ArgoClientDriver) MarkNoDeploy(clusterName string, namespace string, apply bool) error {
	err := a.CheckForRefresh()
	if err != nil {
		return err
	}
	ioProjectCloser, projectServiceClient, err := a.client.NewProjectClient()
	if err != nil {
		log.Printf("Error while fetching the project client %s", err)
		return err
	}
	defer ioProjectCloser.Close()
	listOfProjects, err := projectServiceClient.List(context.TODO(), &project.ProjectQuery{Name: "*"})
	if err != nil {
		log.Printf("Error while fetching the list of projects %s", err)
		return err
	}
	for _, appProject := range listOfProjects.Items {
		for _, dest := range appProject.Spec.Destinations {
			nameMatch, _ := regexp.MatchString(dest.Name, clusterName)
			namespaceMatch, _ := regexp.MatchString(dest.Namespace, namespace)
			if nameMatch && (namespace == "" || namespaceMatch) {
				var namespaceList []string
				var clusterList []string
				if namespace != "" {
					namespaceList = []string{namespace}
				} else {
					clusterList = []string{clusterName}
				}
				windowExists := false
				windowIdx := 0
				for idx, window := range appProject.Spec.SyncWindows {
					//Checking for application list length to be 0 as application level nodeploy is not supported yet
					if window.Kind == "deny" && window.Schedule == "* * * * *" && window.Duration == "720h" &&
						len(window.Applications) == 0 && reflect.DeepEqual(window.Namespaces, namespaceList) &&
						reflect.DeepEqual(window.Clusters, clusterList) && window.ManualSync == false {
						windowExists = true
						windowIdx = idx
						break
					}
				}
				if windowExists && apply {
					break
				} else if windowExists && !apply {
					appProject.Spec.DeleteWindow(windowIdx)
					_, err = projectServiceClient.Update(context.TODO(), &project.ProjectUpdateRequest{Project: &appProject})
					if err != nil {
						log.Printf("Error while deleting sync window to the projects %s", err)
						return err
					}
				} else if !windowExists && apply {
					appProject.Spec.AddWindow("deny", "* * * * *", "720h", []string{}, namespaceList, clusterList, false)
					_, err = projectServiceClient.Update(context.TODO(), &project.ProjectUpdateRequest{Project: &appProject})
					if err != nil {
						log.Printf("Error while adding sync window to the projects %s", err)
						return err
					}
				}
			}
		}
	}
	return nil
}

func (a *ArgoClientDriver) CheckHealthy(argoApplicationName string) bool {
	err := a.CheckForRefresh()
	if err != nil {
		return false
	}
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
