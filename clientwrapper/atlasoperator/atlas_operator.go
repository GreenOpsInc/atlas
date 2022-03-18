package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/argoproj/gitops-engine/pkg/health"
	"github.com/gorilla/mux"
	"github.com/greenopsinc/util/clientrequest"
	"github.com/greenopsinc/util/kubernetesclient"
	"github.com/greenopsinc/util/serializerutil"
	"github.com/greenopsinc/util/starter"
	"github.com/greenopsinc/util/tlsmanager"
	"greenops.io/client/api/generation"
	"greenops.io/client/api/ingest"
	"greenops.io/client/argodriver"
	"greenops.io/client/k8sdriver"
	"greenops.io/client/plugins"
	"greenops.io/client/progressionchecker"
	"greenops.io/client/progressionchecker/datamodel"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	ResourceMetadata clientrequest.DeployResponse //This field is required for an AtlasNonRetryableError
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
var commandDelegatorApi ingest.CommandDelegatorApi
var eventGenerationApi generation.EventGenerationApi

func deploy(request *clientrequest.ClientDeployRequest) (clientrequest.DeployResponse, error) {
	deployType := request.DeployType
	revision := request.RevisionHash
	stringReqBody := request.Payload
	var resourceName string
	var appNamespace string
	var revisionHash string
	var err error
	if deployType == clientrequest.DeployArgoRequest {
		resourceName, appNamespace, revisionHash, err = drivers.argoDriver.Deploy(&stringReqBody, revision)
	} else if deployType == clientrequest.DeployTestRequest {
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
		return clientrequest.DeployResponse{}, AtlasError{
			AtlasErrorType: AtlasRetryableError,
			Err:            err,
		}
	} else {
		return clientrequest.DeployResponse{
			Success:      true,
			ResourceName: resourceName,
			AppNamespace: appNamespace,
			RevisionHash: revisionHash,
		}, nil
	}
}

func deployTaskOrTest(stringReqBody *string) (clientrequest.DeployResponse, error) {
	var resourceName string
	var appNamespace string
	var revisionHash string
	var err error
	var kubernetesCreationRequest clientrequest.KubernetesCreationRequest
	err = json.NewDecoder(strings.NewReader(*stringReqBody)).Decode(&kubernetesCreationRequest)
	if err != nil {
		resourceName, appNamespace, revisionHash = "", "", NotApplicable
	} else {
		if kubernetesCreationRequest.Type == string(plugins.ArgoWorkflow) {
			var plugin plugins.Plugin
			plugin, err = drivers.pluginList.GetPlugin(plugins.ArgoWorkflow)
			if err != nil {
				return clientrequest.DeployResponse{}, AtlasError{
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
		return clientrequest.DeployResponse{}, AtlasError{
			AtlasErrorType: AtlasRetryableError,
			Err:            err,
		}
	} else {
		return clientrequest.DeployResponse{
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
func aggregateResources(request interface{}) (interface{}, error) {
	strongTypeRequest := request.(*clientrequest.ClientAggregateRequest)
	atlasGroup, err := drivers.k8sDriver.Aggregate(strongTypeRequest.ClusterName, strongTypeRequest.Namespace)
	if err != nil {
		return nil, err
	}
	return atlasGroup, nil
}

func getAuditLabel(teamName string, pipelineName string) string {
	label := fmt.Sprintf("%s-%s-stale", teamName, pipelineName)
	return label
}

func labelResources(request interface{}) (interface{}, error) {
	strongTypeRequest := request.(*clientrequest.ClientLabelRequest)
	err := drivers.k8sDriver.Label(strongTypeRequest.GvkResourceList, getAuditLabel(strongTypeRequest.TeamName, strongTypeRequest.PipelineName))
	if err != nil {
		return false, err
	}
	return true, nil
}

func deleteByLabel(request interface{}) (interface{}, error) {
	strongTypeRequest := request.(*clientrequest.ClientDeleteByLabelRequest)
	err := drivers.k8sDriver.DeleteByLabel(getAuditLabel(strongTypeRequest.TeamName, strongTypeRequest.PipelineName), strongTypeRequest.Namespace)
	if err != nil {
		return false, err
	}
	return true, nil
}

func selectiveSyncArgoApp(request *clientrequest.ClientSelectiveSyncRequest) error {
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

func rollbackArgoApp(request *clientrequest.RollbackRequest) (clientrequest.DeployResponse, error) {
	var revisionHash string
	resourceName, appNamespace, revisionHash, err := drivers.argoDriver.Rollback(request.AppName, request.RevisionId)

	if err != nil {
		return clientrequest.DeployResponse{}, AtlasError{
			AtlasErrorType: AtlasRetryableError,
			Err:            err,
		}
	}
	return clientrequest.DeployResponse{
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

func deleteResourceFromConfig(request *clientrequest.ClientDeleteByConfigRequest) error {
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

func deleteApplicationByGvk(request *clientrequest.ClientDeleteByGVKRequest) error {
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

func markNoDeploy(request interface{}) (interface{}, error) {
	//request needs to be of type requestdatatypes.ClientMarkNoDeployRequest
	strongTypeRequest := request.(*clientrequest.ClientMarkNoDeployRequest)
	err := drivers.argoDriver.MarkNoDeploy(strongTypeRequest.ClusterName, strongTypeRequest.Namespace, strongTypeRequest.Apply)
	if err != nil {
		return false, err
	}
	return true, nil
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

func watch(request *clientrequest.WatchRequest) {
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

func deployAndWatch(request *clientrequest.ClientDeployAndWatchRequest) error {
	deployRequest := clientrequest.ClientDeployRequest{
		ClientRequestEventMetadata: request.ClientRequestEventMetadata,
		DeployType:                 request.DeployType,
		RevisionHash:               request.RevisionHash,
		Payload:                    request.Payload,
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
	watchRequest := clientrequest.WatchRequest{
		ClientRequestEventMetadata: request.ClientRequestEventMetadata,
		Type:                       request.WatchType,
		Name:                       deployResponse.ResourceName,
		Namespace:                  deployResponse.AppNamespace,
		PipelineUvn:                request.PipelineUvn,
		TestNumber:                 request.TestNumber,
	}
	watch(&watchRequest)
	return nil
}

func rollbackAndWatch(request *clientrequest.ClientRollbackAndWatchRequest) error {
	rollbackRequest := clientrequest.RollbackRequest{
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
	watchRequest := clientrequest.WatchRequest{
		ClientRequestEventMetadata: request.ClientRequestEventMetadata,
		Type:                       request.WatchType,
		Name:                       deployResponse.ResourceName,
		Namespace:                  deployResponse.AppNamespace,
		PipelineUvn:                request.PipelineUvn,
		TestNumber:                 -1,
	}
	watch(&watchRequest)
	return nil
}

//WARNING: Make sure that the correct request time is sent alongside the function
func handleNotificationRequest(requestId string, f func(interface{}) (interface{}, error), parameter interface{}) {
	var resp interface{}
	var err error
	for i := 0; i < 3; i++ {
		resp, err = f(parameter)
		if err == nil {
			break
		}
	}
	notificationPostProcessing(requestId, resp, err)
}

func handleRequests() {
	var command clientrequest.ClientRequestEvent
	defer func() {
		if err := recover(); err != nil {
			var stronglyTypedError error
			switch err.(type) {
			case error:
				stronglyTypedError = err.(error)
			default:
				stronglyTypedError = errors.New(err.(string))
			}
			log.Printf("Atlas operator exiting with error %s", stronglyTypedError.Error())
			if notificationRequest, ok := command.(clientrequest.NotificationRequestEvent); ok {
				notificationPostProcessing(notificationRequest.GetRequestId(), nil, stronglyTypedError)
			} else {
				requestPostProcessing(command, stronglyTypedError)
			}
		}
	}()
	for {
		var err error
		commands, err := commandDelegatorApi.GetCommands()
		if err != nil {
			log.Printf("Error getting commands %s", err)
			continue
		}
		for _, command = range *commands {
			err = nil
			log.Printf("Handling event type %s", command.GetEvent())
			if command.GetEvent() == serializerutil.ClientDeployRequestType {
				var request *clientrequest.ClientDeployRequest
				request = command.(*clientrequest.ClientDeployRequest)
				var deployResponse clientrequest.DeployResponse
				deployResponse, err = deploy(request)
				if err == nil && deployResponse.Success {
					//If there is no response event (in case of a force deploy), this if will simply pass over event generation
					if request.ResponseEventType != "" && !eventGenerationApi.GenerateResponseEvent(request.ResponseEventType.MakeResponseEvent(&deployResponse, request)) {
						err = AtlasError{AtlasErrorType: AtlasRetryableError, Err: errors.New("response event not generated correctly")}
					}
				}
			} else if command.GetEvent() == serializerutil.ClientDeployAndWatchRequestType {
				var request *clientrequest.ClientDeployAndWatchRequest
				request = command.(*clientrequest.ClientDeployAndWatchRequest)
				err = deployAndWatch(request)
			} else if command.GetEvent() == serializerutil.ClientRollbackAndWatchRequestType {
				var request *clientrequest.ClientRollbackAndWatchRequest
				request = command.(*clientrequest.ClientRollbackAndWatchRequest)
				err = rollbackAndWatch(request)
			} else if command.GetEvent() == serializerutil.ClientDeleteByGvkRequestType {
				var request *clientrequest.ClientDeleteByGVKRequest
				request = command.(*clientrequest.ClientDeleteByGVKRequest)
				err = deleteApplicationByGvk(request)
			} else if command.GetEvent() == serializerutil.ClientDeleteByConfigRequestType {
				var request *clientrequest.ClientDeleteByConfigRequest
				request = command.(*clientrequest.ClientDeleteByConfigRequest)
				err = deleteResourceFromConfig(request)
			} else if command.GetEvent() == serializerutil.ClientSelectiveSyncRequestType {
				var request *clientrequest.ClientSelectiveSyncRequest
				request = command.(*clientrequest.ClientSelectiveSyncRequest)
				err = selectiveSyncArgoApp(request)
			} else if command.GetEvent() == serializerutil.ClientMarkNoDeployRequestType {
				var request *clientrequest.ClientMarkNoDeployRequest
				request = command.(*clientrequest.ClientMarkNoDeployRequest)
				handleNotificationRequest(request.RequestId, markNoDeploy, request)
				continue
			} else if command.GetEvent() == serializerutil.ClientAggregateRequestType {
				var request *clientrequest.ClientAggregateRequest
				request = command.(*clientrequest.ClientAggregateRequest)
				handleNotificationRequest(request.RequestId, aggregateResources, request)
				continue
			} else if command.GetEvent() == serializerutil.ClientLabelRequestType {
				var request *clientrequest.ClientLabelRequest
				request = command.(*clientrequest.ClientLabelRequest)
				handleNotificationRequest(request.RequestId, labelResources, request)
				continue
			} else if command.GetEvent() == serializerutil.ClientDeleteByLabelRequestType {
				var request *clientrequest.ClientDeleteByLabelRequest
				request = command.(*clientrequest.ClientDeleteByLabelRequest)
				handleNotificationRequest(request.RequestId, deleteByLabel, request)
				continue
			}
			requestPostProcessing(command, err)
		}
		duration := 10 * time.Second
		time.Sleep(duration)
	}
}

func notificationPostProcessing(requestId string, resp interface{}, err error) {
	generationSuccess := false
	for !generationSuccess {
		var notification generation.Notification
		if err != nil {
			notification.Successful = false
			notification.Body = err.Error()
			generationSuccess = eventGenerationApi.GenerateNotification(requestId, notification)
		} else {
			notification.Successful = true
			notification.Body = resp
			generationSuccess = eventGenerationApi.GenerateNotification(requestId, notification)
		}
		if !generationSuccess {
			log.Printf("Caught error trying to send notification...looping until request can be sent correctly")
		}
	}
	generationSuccess = false
	for !generationSuccess {
		retryError := commandDelegatorApi.AckHeadOfNotificationList()
		if retryError != nil {
			log.Printf("Caught error trying to send ack notification message: %s...looping until request can be sent correctly", retryError)
			generationSuccess = false
			continue
		}
		generationSuccess = true
	}
}

func requestPostProcessing(command clientrequest.ClientRequestEvent, err error) {
	generationSuccess := false
	for !generationSuccess {
		var retryError error
		if err != nil {
			atlasError, ok := err.(AtlasError)
			//If the error is not an AtlasError, it should be treated as a RetryableError
			if !ok {
				atlasError = AtlasError{
					AtlasErrorType: AtlasRetryableError,
					Err:            err,
				}
			}

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
				failureEvent := datamodel.MakeFailureEventEvent(command.GetClientMetadata(), clientrequest.DeployResponse{}, "", atlasError.Error())

				generationSuccess = eventGenerationApi.GenerateEvent(failureEvent)
				if !generationSuccess {
					log.Printf("Unable to send failure notification...looping until notification can be sent correctly")
					continue
				}
			}
		}
		retryError = commandDelegatorApi.AckHeadOfRequestList()
		if retryError != nil {
			log.Printf("Caught error trying to send ack request message: %s...looping until request can be sent correctly", retryError)
			err = nil
			generationSuccess = false
			continue
		}
		generationSuccess = true
	}
}

func main() {
	var err error
	kubernetesDriver := k8sdriver.New()
	var tm tlsmanager.Manager
	kubernetesClient := kubernetesclient.New()
	if starter.GetNoAuthClientConfig() {
		tm = tlsmanager.NoAuth()
	} else {
		tm = tlsmanager.New(kubernetesClient)
	}
	argoDriver := argodriver.New(&kubernetesDriver, tm)
	commandDelegatorApi, err = ingest.Create(argoDriver.(argodriver.ArgoAuthClient), tm)
	if err != nil {
		log.Fatal("command delegator API setup failed: ", err.Error())
	}
	eventGenerationApi, err = generation.Create(argoDriver.(argodriver.ArgoAuthClient), kubernetesClient, tm)
	if err != nil {
		log.Fatal("event generation API setup failed: ", err.Error())
	}

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
	for {
		handleRequests()
	}
}
