package main

import (
	"encoding/json"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/argoproj/gitops-engine/pkg/health"
	"github.com/gorilla/mux"
	"greenops.io/client/argodriver"
	"greenops.io/client/atlasoperator/requestdatatypes"
	"greenops.io/client/k8sdriver"
	"greenops.io/client/progressionchecker"
	"greenops.io/client/progressionchecker/datamodel"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"log"
	"net/http"
	"strings"
)

const (
	NotApplicable string = "NotApplicable"
)

type Drivers struct {
	k8sDriver  k8sdriver.KubernetesClient
	argoDriver argodriver.ArgoClient
}

var drivers Drivers
var channel chan string

func deploy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deployType := vars["type"]
	revision := vars["revision"]
	byteReqBody, _ := ioutil.ReadAll(r.Body)
	stringReqBody := string(byteReqBody)
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
		success, resourceName, appNamespace = drivers.k8sDriver.Deploy(&stringReqBody)
		revisionHash = NotApplicable
	}

	json.NewEncoder(w).Encode(
		requestdatatypes.DeployResponse{
			Success:      success,
			ResourceName: resourceName,
			AppNamespace: appNamespace,
			RevisionHash: revisionHash,
		},
	)
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

func selectiveSyncArgoApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamName := vars["teamName"]
	pipelineName := vars["pipelineName"]
	stepName := vars["stepName"]
	appName := vars["appName"]
	stringRevisionHash := vars["revisionId"]
	var gvkGroupRequest requestdatatypes.GvkGroupRequest
	byteReqBody, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(byteReqBody, &gvkGroupRequest)

	success, resourceName, appNamespace, revisionHash := drivers.argoDriver.SelectiveSync(appName, stringRevisionHash, gvkGroupRequest)
	if success {
		key := makeWatchKeyFromRequest(datamodel.WatchArgoApplicationKey, vars["orgName"], teamName, pipelineName, stepName, -1, resourceName, appNamespace)
		byteKey, _ := json.Marshal(key)
		channel <- string(byteKey)
		channel <- progressionchecker.EndTransactionMarker
	}

	json.NewEncoder(w).Encode(
		requestdatatypes.RemediationResponse{
			Success:      success,
			ResourceName: resourceName,
			AppNamespace: appNamespace,
			RevisionHash: revisionHash,
		})
}

func rollbackArgoApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appName := vars["appName"]
	stringRevisionHash := vars["revisionId"]
	var revisionHash string
	success, resourceName, appNamespace, revisionHash := drivers.argoDriver.Rollback(appName, stringRevisionHash)

	json.NewEncoder(w).Encode(
		requestdatatypes.DeployResponse{
			Success:      success,
			ResourceName: resourceName,
			AppNamespace: appNamespace,
			RevisionHash: revisionHash,
		})
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

func deleteResourceFromConfig(w http.ResponseWriter, r *http.Request) {
	//vars := mux.Vars(r)
	//deleteType := vars["type"]
	byteReqBody, _ := ioutil.ReadAll(r.Body)
	stringReqBody := string(byteReqBody)
	success := drivers.k8sDriver.DeleteBasedOnConfig(&stringReqBody)

	json.NewEncoder(w).Encode(success)
}

func deleteApplication(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupVersionKind := schema.GroupVersionKind{
		Group:   vars["group"],
		Version: vars["version"],
		Kind:    vars["kind"],
	}
	applicationName := vars["name"]
	var success bool
	if strings.Contains(groupVersionKind.Group, "argo") {
		success = drivers.argoDriver.Delete(applicationName)
	} else {
		panic("K8S delete method not implemented yet.")
	}

	json.NewEncoder(w).Encode(success)
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

func watch(w http.ResponseWriter, r *http.Request) {
	var watchRequest requestdatatypes.WatchRequest
	err := json.NewDecoder(r.Body).Decode(&watchRequest)
	if err != nil {
		json.NewEncoder(w).Encode(false)
		return
	}
	vars := mux.Vars(r)
	var watchKeyType datamodel.WatchKeyType
	if watchRequest.Type == string(datamodel.WatchTestKey) {
		watchKeyType = datamodel.WatchTestKey
	} else {
		watchKeyType = datamodel.WatchArgoApplicationKey
	}

	key := makeWatchKeyFromRequest(watchKeyType, vars["orgName"], watchRequest.TeamName, watchRequest.PipelineName, watchRequest.StepName, watchRequest.TestNumber, watchRequest.Name, watchRequest.Namespace)
	byteKey, _ := json.Marshal(key)
	channel <- string(byteKey)
	channel <- progressionchecker.EndTransactionMarker
	json.NewEncoder(w).Encode(true)
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

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/deploy/{orgName}/{type}/{revision}", deploy).Methods("POST")
	myRouter.HandleFunc("/deploy/{orgName}/{type}/{argoAppName}", deployArgoAppByName).Methods("POST")
	myRouter.HandleFunc("/delete/{orgName}/{type}/{resourceName}/{resourceNamespace}/{group}/{version}{kind}", deleteResource).Methods("POST")
	myRouter.HandleFunc("/delete/{orgName}/{type}", deleteResourceFromConfig).Methods("POST")
	myRouter.HandleFunc("/rollback/{orgName}/{appName}/{revisionId}", rollbackArgoApp).Methods("POST")
	myRouter.HandleFunc("/sync/{orgName}/{teamName}/{pipelineName}/{stepName}/{appName}/{revisionId}", selectiveSyncArgoApp).Methods("POST")
	myRouter.HandleFunc("/delete/{group}/{version}/{kind}/{name}", deleteApplication).Methods("POST")
	myRouter.HandleFunc("/checkStatus/{group}/{version}/{kind}/{name}", checkStatus).Methods("GET")
	myRouter.HandleFunc("/watch/{orgName}", watch).Methods("POST")
	log.Fatal(http.ListenAndServe(":9091", myRouter))
}

func main() {
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
	go progressionchecker.Start(getRestrictedKubernetesClient, getRestrictedArgoClient, channel)
	handleRequests()
}
