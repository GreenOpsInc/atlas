package main

import (
	"encoding/json"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/argoproj/gitops-engine/pkg/health"
	"github.com/gorilla/mux"
	"greenops.io/client/api/generation"
	"greenops.io/client/api/ingest"
	"greenops.io/client/argodriver"
	"greenops.io/client/atlasoperator/requestdatatypes"
	"greenops.io/client/k8sdriver"
	"greenops.io/client/progressionchecker"
	"greenops.io/client/progressionchecker/datamodel"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	NotApplicable string = "NotApplicable"
)

type Drivers struct {
	k8sDriver  k8sdriver.KubernetesClient
	argoDriver argodriver.ArgoClient
}

type AtlasErrorType string

const (
	AtlasRetryableError    AtlasErrorType = "AtlasRetryableError"
	AtlasNonRetryableError AtlasErrorType = "AtlasNonRetryableError"
)

type AtlasError struct {
	AtlasErrorType   AtlasErrorType
	ResourceMetadata requestdatatypes.DeployResponse //This field is required for an AtlasNonRetryableError
	Err              error
}

func (a AtlasError) Error() string {
	return a.Err.Error()
}

var drivers Drivers
var channel chan string

func deploy(request *requestdatatypes.ClientDeployRequest) (requestdatatypes.DeployResponse, error) {
	deployType := request.DeployType
	revision := request.RevisionHash
	stringReqBody := request.Payload
	var success bool
	var resourceName string
	var appNamespace string
	var revisionHash string
	if deployType == requestdatatypes.DeployArgoRequest {
		success, resourceName, appNamespace, revisionHash = drivers.argoDriver.Deploy(&stringReqBody, revision)
	} else if deployType == requestdatatypes.DeployTestRequest {
		var kubernetesCreationRequest requestdatatypes.KubernetesCreationRequest
		err := json.NewDecoder(strings.NewReader(stringReqBody)).Decode(&kubernetesCreationRequest)
		if err != nil {
			success, resourceName, appNamespace, revisionHash = false, "", "", NotApplicable
		} else {
			success, resourceName, appNamespace = drivers.k8sDriver.CreateAndDeploy(
				kubernetesCreationRequest.Kind,
				kubernetesCreationRequest.ObjectName,
				kubernetesCreationRequest.Namespace,
				kubernetesCreationRequest.ImageName,
				kubernetesCreationRequest.Command,
				kubernetesCreationRequest.Args,
				kubernetesCreationRequest.Config,
				kubernetesCreationRequest.Variables,
			)
			revisionHash = NotApplicable
		}
	} else {
		resources := strings.Split(stringReqBody, "---")
		for _, resource := range resources {
			success, resourceName, appNamespace = drivers.k8sDriver.Deploy(&resource)
			revisionHash = NotApplicable
			if !success {
				break
			}
		}
	}

	if success {
		return requestdatatypes.DeployResponse{
			Success:      success,
			ResourceName: resourceName,
			AppNamespace: appNamespace,
			RevisionHash: revisionHash,
		}, nil
	} else {
		return requestdatatypes.DeployResponse{}, AtlasError{
			AtlasErrorType: AtlasRetryableError,
			Err:            nil,
		}
	}
}

func deployArgoAppByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	//requestType := vars["type"]
	argoAppName := vars["argoAppName"]
	var success bool
	var resourceName string
	var appNamespace string
	var revisionHash string
	success, resourceName, appNamespace, revisionHash = drivers.argoDriver.Sync(argoAppName)

	json.NewEncoder(w).Encode(
		requestdatatypes.DeployResponse{
			Success:      success,
			ResourceName: resourceName,
			AppNamespace: appNamespace,
			RevisionHash: revisionHash,
		},
	)
}

func selectiveSyncArgoApp(request *requestdatatypes.ClientSelectiveSyncRequest) error {
	success, resourceName, appNamespace, _ := drivers.argoDriver.SelectiveSync(request.AppName, request.RevisionHash, request.GvkResourceList)
	if success {
		key := makeWatchKeyFromRequest(
			datamodel.WatchArgoApplicationKey,
			request.OrgName,
			request.TeamName,
			request.PipelineName,
			request.StepName,
			-1,
			resourceName,
			appNamespace,
		)
		byteKey, _ := json.Marshal(key)
		channel <- string(byteKey)
		channel <- progressionchecker.EndTransactionMarker
		return nil
	} else {
		return AtlasError{
			AtlasErrorType: AtlasRetryableError,
			Err:            nil,
		}
	}
}

func rollbackArgoApp(request *requestdatatypes.RollbackRequest) (requestdatatypes.DeployResponse, error) {
	var revisionHash string
	success, resourceName, appNamespace, revisionHash := drivers.argoDriver.Rollback(request.AppName, request.RevisionId)

	if !success {
		return requestdatatypes.DeployResponse{}, AtlasError{
			AtlasErrorType: AtlasRetryableError,
			Err:            nil,
		}
	}
	return requestdatatypes.DeployResponse{
		Success:      success,
		ResourceName: resourceName,
		AppNamespace: appNamespace,
		RevisionHash: revisionHash,
	}, nil
}

func deleteResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	//deleteType := vars["type"]
	resourceName := vars["resourceName"]
	resourceNamespace := vars["resourceNamespace"]
	resourceGroup := vars["group"]
	resourceVersion := vars["version"]
	resourceKind := vars["kind"]
	success := drivers.k8sDriver.Delete(resourceName, resourceNamespace, schema.GroupVersionKind{
		Group:   resourceGroup,
		Version: resourceVersion,
		Kind:    resourceKind,
	})

	json.NewEncoder(w).Encode(success)
}

func deleteResourceFromConfig(request *requestdatatypes.ClientDeleteByConfigRequest) error {
	resources := strings.Split(request.ConfigPayload, "---")
	var success bool
	for _, resource := range resources {
		success = drivers.k8sDriver.DeleteBasedOnConfig(&resource)
		if !success {
			return AtlasError{
				AtlasErrorType: AtlasRetryableError,
				Err:            nil,
			}
		}
	}
	return nil
}

func deleteApplicationByGvk(request *requestdatatypes.ClientDeleteByGvkRequest) error {
	groupVersionKind := schema.GroupVersionKind{
		Group:   request.Group,
		Version: request.Version,
		Kind:    request.Kind,
	}
	applicationName := request.ResourceName
	var success bool
	if strings.Contains(groupVersionKind.Group, "argo") {
		success = drivers.argoDriver.Delete(applicationName)
	} else {
		success = drivers.k8sDriver.Delete(request.ResourceName, request.ResourceNamespace, groupVersionKind)
	}

	if !success {
		return AtlasError{
			AtlasErrorType: AtlasRetryableError,
			Err:            nil,
		}
	}
	return nil
}

func checkStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupVersionKind := schema.GroupVersionKind{
		Group:   vars["group"],
		Version: vars["version"],
		Kind:    vars["kind"],
	}
	applicationName := vars["name"]
	var success bool
	if strings.Contains(groupVersionKind.Group, "argo") {
		success = drivers.argoDriver.CheckHealthy(applicationName)
	} else {
		panic("K8S health check not implemented yet.")
	}

	json.NewEncoder(w).Encode(success)
}

func watch(request *requestdatatypes.WatchRequest) {
	var watchKeyType datamodel.WatchKeyType
	if request.Type == string(datamodel.WatchTestKey) {
		watchKeyType = datamodel.WatchTestKey
	} else {
		watchKeyType = datamodel.WatchArgoApplicationKey
	}

	key := datamodel.WatchKey{
		WatchKeyMetaData: datamodel.WatchKeyMetaData{
			Type:         watchKeyType,
			OrgName:      request.OrgName,
			TeamName:     request.TeamName,
			PipelineName: request.PipelineName,
			StepName:     request.StepName,
			TestNumber:   request.TestNumber,
		},
		Name:                     request.Name,
		Namespace:                request.Namespace,
		HealthStatus:             string(health.HealthStatusMissing),
		SyncStatus:               string(v1alpha1.SyncStatusCodeOutOfSync),
		GeneratedCompletionEvent: false,
	}
	byteKey, _ := json.Marshal(key)
	channel <- string(byteKey)
	channel <- progressionchecker.EndTransactionMarker
}

func makeWatchKeyFromRequest(watchKeyType datamodel.WatchKeyType, orgName string, teamName string, pipelineName string, stepName string, testNumber int, name string, namespace string) *datamodel.WatchKey {
	return &datamodel.WatchKey{
		WatchKeyMetaData: datamodel.WatchKeyMetaData{
			Type:         watchKeyType,
			OrgName:      orgName,
			TeamName:     teamName,
			PipelineName: pipelineName,
			StepName:     stepName,
			TestNumber:   testNumber,
		},
		Name:                     name,
		Namespace:                namespace,
		HealthStatus:             string(health.HealthStatusMissing),
		SyncStatus:               string(v1alpha1.SyncStatusCodeOutOfSync),
		GeneratedCompletionEvent: false,
	}
}

func deployAndWatch(request *requestdatatypes.ClientDeployAndWatchRequest) error {
	deployRequest := requestdatatypes.ClientDeployRequest{
		ClientEventMetadata: request.ClientEventMetadata,
		DeployType:          request.DeployType,
		RevisionHash:        request.RevisionHash,
		Payload:             request.Payload,
	}
	deployResponse, err := deploy(&deployRequest)
	if err != nil {
		return AtlasError{
			AtlasErrorType: AtlasRetryableError,
			Err:            nil,
		}
	}
	watchRequest := requestdatatypes.WatchRequest{
		ClientEventMetadata: request.ClientEventMetadata,
		Type:                request.WatchType,
		Name:                deployResponse.ResourceName,
		Namespace:           deployResponse.AppNamespace,
		TestNumber:          request.TestNumber,
	}
	watch(&watchRequest)
	return nil
}

func rollbackAndWatch(request *requestdatatypes.ClientRollbackAndWatchRequest) error {
	rollbackRequest := requestdatatypes.RollbackRequest{
		AppName:    request.AppName,
		RevisionId: request.RevisionHash,
	}
	deployResponse, err := rollbackArgoApp(&rollbackRequest)
	if err != nil {
		return AtlasError{
			AtlasErrorType: AtlasRetryableError,
			Err:            nil,
		}
	}
	watchRequest := requestdatatypes.WatchRequest{
		ClientEventMetadata: request.ClientEventMetadata,
		Type:                request.WatchType,
		Name:                deployResponse.ResourceName,
		Namespace:           deployResponse.AppNamespace,
		TestNumber:          -1,
	}
	watch(&watchRequest)
	return nil
}

func handleRequests(commandDelegatorApi ingest.CommandDelegatorApi, eventGenerationApi generation.EventGenerationApi) {
	for {
		var err error
		commands, err := commandDelegatorApi.GetCommands()
		if err != nil {
			log.Printf("Error getting commands %s", err)
			continue
		}
		for _, command := range *commands {
			err = nil
			log.Printf("Handling event type %s", command.GetEvent())
			if command.GetEvent() == requestdatatypes.ClientDeployRequestType {
				var request requestdatatypes.ClientDeployRequest
				request = command.(requestdatatypes.ClientDeployRequest)
				var deployResponse requestdatatypes.DeployResponse
				deployResponse, err = deploy(&request)
				if deployResponse.Success {
					eventGenerationApi.GenerateResponseEvent(request.ResponseEventType.MakeResponseEvent(&deployResponse, &request))
				}
			} else if command.GetEvent() == requestdatatypes.ClientDeployAndWatchRequestType {
				var request requestdatatypes.ClientDeployAndWatchRequest
				request = command.(requestdatatypes.ClientDeployAndWatchRequest)
				err = deployAndWatch(&request)
			} else if command.GetEvent() == requestdatatypes.ClientRollbackAndWatchRequestType {
				var request requestdatatypes.ClientRollbackAndWatchRequest
				request = command.(requestdatatypes.ClientRollbackAndWatchRequest)
				err = rollbackAndWatch(&request)
			} else if command.GetEvent() == requestdatatypes.ClientDeleteByGvkRequestType {
				var request requestdatatypes.ClientDeleteByGvkRequest
				request = command.(requestdatatypes.ClientDeleteByGvkRequest)
				err = deleteApplicationByGvk(&request)
			} else if command.GetEvent() == requestdatatypes.ClientDeleteByConfigRequestType {
				var request requestdatatypes.ClientDeleteByConfigRequest
				request = command.(requestdatatypes.ClientDeleteByConfigRequest)
				err = deleteResourceFromConfig(&request)
			} else if command.GetEvent() == requestdatatypes.ClientSelectiveSyncRequestType {
				var request requestdatatypes.ClientSelectiveSyncRequest
				request = command.(requestdatatypes.ClientSelectiveSyncRequest)
				err = selectiveSyncArgoApp(&request)
			}
			if err != nil {
				atlasError := err.(AtlasError)
				if atlasError.AtlasErrorType == AtlasRetryableError {
					//TODO: The command should be "postponed". Essentially acked and then re-inserted at the end of the list
					continue
				} else {
					failureEvent := datamodel.MakeFailureEventEvent(command.GetClientMetadata(), requestdatatypes.DeployResponse{}, "", atlasError.Error())
					generationSuccess := false
					for i := 0; i < progressionchecker.HttpRequestRetryLimit; i++ {
						generationSuccess = eventGenerationApi.GenerateEvent(failureEvent)
						if generationSuccess {
							break
						}
					}
					if !generationSuccess {
						continue
					}
				}
			}
			commandDelegatorApi.AckHeadOfRequestList()
		}
		duration := 10 * time.Second
		time.Sleep(duration)
	}
}

func main() {
	commandDelegatorApi := ingest.Create()
	eventGenerationApi := generation.Create()
	kubernetesDriver := k8sdriver.New()
	argoDriver := argodriver.New(&kubernetesDriver)
	drivers = Drivers{
		k8sDriver:  kubernetesDriver,
		argoDriver: argoDriver,
	}
	channel = make(chan string)
	var getRestrictedKubernetesClient k8sdriver.KubernetesClientGetRestricted
	getRestrictedKubernetesClient = kubernetesDriver
	var getRestrictedArgoClient argodriver.ArgoGetRestrictedClient
	getRestrictedArgoClient = argoDriver
	go progressionchecker.Start(getRestrictedKubernetesClient, getRestrictedArgoClient, eventGenerationApi, channel)
	handleRequests(commandDelegatorApi, eventGenerationApi)
}
