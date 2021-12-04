package main

import (
	"encoding/json"
	"errors"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/argoproj/gitops-engine/pkg/health"
	"github.com/gorilla/mux"
	"greenops.io/client/api/generation"
	"greenops.io/client/api/ingest"
	"greenops.io/client/argodriver"
	"greenops.io/client/atlasoperator/requestdatatypes"
	"greenops.io/client/k8sdriver"
	"greenops.io/client/plugins"
	"greenops.io/client/progressionchecker"
	"greenops.io/client/progressionchecker/datamodel"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	ArgoWorkflowKind string = "Workflow"
	NotApplicable    string = "NotApplicable"
)

type Drivers struct {
	k8sDriver  k8sdriver.KubernetesClient
	argoDriver argodriver.ArgoClient
	pluginList plugins.Plugins
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
	if a.Err == nil {
		return "No error message available"
	}
	return a.Err.Error()
}

var drivers Drivers
var channel chan string

func deploy(request *requestdatatypes.ClientDeployRequest) (requestdatatypes.DeployResponse, error) {
	deployType := request.DeployType
	revision := request.RevisionHash
	stringReqBody := request.Payload
	var resourceName string
	var appNamespace string
	var revisionHash string
	var err error
	if deployType == requestdatatypes.DeployArgoRequest {
		resourceName, appNamespace, revisionHash, err = drivers.argoDriver.Deploy(&stringReqBody, revision)
	} else if deployType == requestdatatypes.DeployTestRequest {
		return deployTaskOrTest(&stringReqBody)
	} else {
		resources := strings.Split(stringReqBody, "---")
		for _, resource := range resources {
			resourceName, appNamespace, err = drivers.k8sDriver.Deploy(&resource)
			revisionHash = NotApplicable
			if err != nil {
				break
			}
		}
	}

	if err != nil {
		return requestdatatypes.DeployResponse{}, AtlasError{
			AtlasErrorType: AtlasRetryableError,
			Err:            err,
		}
	} else {
		return requestdatatypes.DeployResponse{
			Success:      true,
			ResourceName: resourceName,
			AppNamespace: appNamespace,
			RevisionHash: revisionHash,
		}, nil
	}
}

func deployTaskOrTest(stringReqBody *string) (requestdatatypes.DeployResponse, error) {
	var resourceName string
	var appNamespace string
	var revisionHash string
	var err error
	var kubernetesCreationRequest requestdatatypes.KubernetesCreationRequest
	err = json.NewDecoder(strings.NewReader(*stringReqBody)).Decode(&kubernetesCreationRequest)
	if err != nil {
		resourceName, appNamespace, revisionHash = "", "", NotApplicable
	} else {
		if kubernetesCreationRequest.Type == string(plugins.ArgoWorkflow) {
			var plugin plugins.Plugin
			plugin, err = drivers.pluginList.GetPlugin(plugins.ArgoWorkflow)
			if err != nil {
				return requestdatatypes.DeployResponse{}, AtlasError{
					AtlasErrorType: AtlasRetryableError,
					Err:            err,
				}
			}
			resourceName, appNamespace, err = plugin.PluginObject.CreateAndDeploy(&kubernetesCreationRequest.Config, kubernetesCreationRequest.Variables)
		} else {
			resourceName, appNamespace, err = drivers.k8sDriver.CreateAndDeploy(
				kubernetesCreationRequest.Kind,
				kubernetesCreationRequest.ObjectName,
				kubernetesCreationRequest.Namespace,
				kubernetesCreationRequest.ImageName,
				kubernetesCreationRequest.Command,
				kubernetesCreationRequest.Args,
				kubernetesCreationRequest.Config,
				kubernetesCreationRequest.VolumeFilename,
				kubernetesCreationRequest.VolumeConfig,
				kubernetesCreationRequest.Variables,
			)
		}
		revisionHash = NotApplicable
	}

	if err != nil {
		return requestdatatypes.DeployResponse{}, AtlasError{
			AtlasErrorType: AtlasRetryableError,
			Err:            err,
		}
	} else {
		return requestdatatypes.DeployResponse{
			Success:      true,
			ResourceName: resourceName,
			AppNamespace: appNamespace,
			RevisionHash: revisionHash,
		}, nil
	}
}

//func deployArgoAppByName(w http.ResponseWriter, r *http.Request) {
//	vars := mux.Vars(r)
//	//requestType := vars["type"]
//	argoAppName := vars["argoAppName"]
//	var resourceName string
//	var appNamespace string
//	var revisionHash string
//	resourceName, appNamespace, revisionHash, _ = drivers.argoDriver.Sync(argoAppName)
//
//	json.NewEncoder(w).Encode(
//		requestdatatypes.DeployResponse{
//			Success:      true,
//			ResourceName: resourceName,
//			AppNamespace: appNamespace,
//			RevisionHash: revisionHash,
//		},
//	)
//}

func selectiveSyncArgoApp(request *requestdatatypes.ClientSelectiveSyncRequest) error {
	resourceName, appNamespace, _, err := drivers.argoDriver.SelectiveSync(request.AppName, request.RevisionHash, request.GvkResourceList)
	if err == nil {
		key := makeWatchKeyFromRequest(
			datamodel.WatchArgoApplicationKey,
			request.OrgName,
			request.TeamName,
			request.PipelineName,
			request.StepName,
			-1,
			resourceName,
			appNamespace,
			request.PipelineUvn,
		)
		byteKey, _ := json.Marshal(key)
		channel <- string(byteKey)
		channel <- progressionchecker.EndTransactionMarker
		return nil
	} else {
		return AtlasError{
			AtlasErrorType: AtlasRetryableError,
			Err:            err,
		}
	}
}

func rollbackArgoApp(request *requestdatatypes.RollbackRequest) (requestdatatypes.DeployResponse, error) {
	var revisionHash string
	resourceName, appNamespace, revisionHash, err := drivers.argoDriver.Rollback(request.AppName, request.RevisionId)

	if err != nil {
		return requestdatatypes.DeployResponse{}, AtlasError{
			AtlasErrorType: AtlasRetryableError,
			Err:            err,
		}
	}
	return requestdatatypes.DeployResponse{
		Success:      true,
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
	var err error
	for _, resource := range resources {
		err = drivers.k8sDriver.DeleteBasedOnConfig(&resource)
		if err != nil {
			return AtlasError{
				AtlasErrorType: AtlasRetryableError,
				Err:            err,
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
	var err error
	if strings.Contains(groupVersionKind.Group, "argo") {
		err = drivers.argoDriver.Delete(applicationName)
	} else {
		err = drivers.k8sDriver.Delete(request.ResourceName, request.ResourceNamespace, groupVersionKind)
	}

	if err != nil {
		return AtlasError{
			AtlasErrorType: AtlasRetryableError,
			Err:            err,
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
	} else if request.Type == string(datamodel.WatchArgoWorkflowKey) {
		watchKeyType = datamodel.WatchArgoWorkflowKey
	} else {
		watchKeyType = datamodel.WatchArgoApplicationKey
	}

	key := datamodel.WatchKey{
		WatchKeyMetaData: datamodel.WatchKeyMetaData{
			Type:         watchKeyType,
			OrgName:      request.OrgName,
			TeamName:     request.TeamName,
			PipelineName: request.PipelineName,
			PipelineUvn:  request.PipelineUvn,
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

func makeWatchKeyFromRequest(watchKeyType datamodel.WatchKeyType, orgName string, teamName string, pipelineName string, stepName string, testNumber int, name string, namespace string, pipelineUvn string) *datamodel.WatchKey {
	return &datamodel.WatchKey{
		WatchKeyMetaData: datamodel.WatchKeyMetaData{
			Type:         watchKeyType,
			OrgName:      orgName,
			TeamName:     teamName,
			PipelineName: pipelineName,
			PipelineUvn:  pipelineUvn,
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
		switch err.(type) {
		case AtlasError:
			return err
		default:
			return AtlasError{
				AtlasErrorType: AtlasRetryableError,
				Err:            err,
			}
		}
	}
	watchRequest := requestdatatypes.WatchRequest{
		ClientEventMetadata: request.ClientEventMetadata,
		Type:                request.WatchType,
		Name:                deployResponse.ResourceName,
		Namespace:           deployResponse.AppNamespace,
		PipelineUvn:         request.PipelineUvn,
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
		switch err.(type) {
		case AtlasError:
			return err
		default:
			return AtlasError{
				AtlasErrorType: AtlasRetryableError,
				Err:            err,
			}
		}
	}
	watchRequest := requestdatatypes.WatchRequest{
		ClientEventMetadata: request.ClientEventMetadata,
		Type:                request.WatchType,
		Name:                deployResponse.ResourceName,
		Namespace:           deployResponse.AppNamespace,
		PipelineUvn:         request.PipelineUvn,
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
				if err == nil && deployResponse.Success {
					if !eventGenerationApi.GenerateResponseEvent(request.ResponseEventType.MakeResponseEvent(&deployResponse, &request)) {
						err = AtlasError{AtlasErrorType: AtlasRetryableError, Err: errors.New("response event not generated correctly")}
					}
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
			generationSuccess := false
			for !generationSuccess {
				var retryError error
				if err != nil {
					atlasError := err.(AtlasError)
					if atlasError.AtlasErrorType == AtlasRetryableError && !command.GetClientMetadata().FinalTry {
						log.Printf("Caught retryable error with message %s", atlasError)
						retryError = commandDelegatorApi.RetryRequest()
						if retryError != nil {
							log.Printf("Caught error trying to send request for retry: %s...looping until request can be sent correctly", retryError)
							generationSuccess = false
							continue
						}
						generationSuccess = true
						//Retry automatically acks the head of the request list, no need to do it again
						break
					} else {
						log.Printf("Caught non-retryable error with message %s", atlasError)
						failureEvent := datamodel.MakeFailureEventEvent(command.GetClientMetadata(), requestdatatypes.DeployResponse{}, "", atlasError.Error())

						generationSuccess = eventGenerationApi.GenerateEvent(failureEvent)
						if !generationSuccess {
							log.Printf("Unable to send failure notification...looping until notification can be sent correctly")
							continue
						}
					}
				}
				retryError = commandDelegatorApi.AckHeadOfRequestList()
				if retryError != nil {
					log.Printf("Caught error trying to send ack message: %s...looping until request can be sent correctly", retryError)
					err = nil
					generationSuccess = false
					continue
				}
				generationSuccess = true
			}
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
	var pluginList plugins.Plugins
	pluginList = make([]plugins.Plugin, 0)
	drivers = Drivers{
		k8sDriver:  kubernetesDriver,
		argoDriver: argoDriver,
		pluginList: pluginList,
	}
	channel = make(chan string)
	var getRestrictedKubernetesClient k8sdriver.KubernetesClientGetRestricted
	getRestrictedKubernetesClient = kubernetesDriver
	var getRestrictedArgoClient argodriver.ArgoGetRestrictedClient
	getRestrictedArgoClient = argoDriver
	go progressionchecker.Start(getRestrictedKubernetesClient, getRestrictedArgoClient, eventGenerationApi, pluginList, channel)
	handleRequests(commandDelegatorApi, eventGenerationApi)
}
