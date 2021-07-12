package main

import (
	"encoding/json"
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

type Drivers struct {
	k8sDriver  k8sdriver.KubernetesClient
	argoDriver argodriver.ArgoClient
}

var drivers Drivers
var channel chan string

func deploy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deployType := vars["type"]
	byteReqBody, _ := ioutil.ReadAll(r.Body)
	stringReqBody := string(byteReqBody)
	var success bool
	var appNamespace string
	if deployType == requestdatatypes.DeployArgoRequest {
		success, appNamespace = drivers.argoDriver.Deploy(stringReqBody)
	} else if deployType == requestdatatypes.DeployTestRequest {
		var kubernetesCreationRequest requestdatatypes.KubernetesCreationRequest
		err := json.NewDecoder(strings.NewReader(stringReqBody)).Decode(&kubernetesCreationRequest)
		if err != nil {
			success, appNamespace = false, ""
		} else {
			success, appNamespace = drivers.k8sDriver.CreateAndDeploy(
				kubernetesCreationRequest.Kind,
				kubernetesCreationRequest.ObjectName,
				kubernetesCreationRequest.Namespace,
				kubernetesCreationRequest.ImageName,
				kubernetesCreationRequest.Command,
				kubernetesCreationRequest.Args,
				kubernetesCreationRequest.Variables,
			)
		}
	} else {
		success, appNamespace = drivers.k8sDriver.Deploy(stringReqBody)
	}

	json.NewEncoder(w).Encode(
		requestdatatypes.DeployResponse{
			Success:      success,
			AppNamespace: appNamespace,
		})
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

	key := datamodel.WatchKey{
		WatchKeyMetaData: datamodel.WatchKeyMetaData{
			Type:         watchKeyType,
			OrgName:      vars["orgName"],
			TeamName:     watchRequest.TeamName,
			PipelineName: watchRequest.PipelineName,
			StepName:     watchRequest.StepName,
		},
		Name:                     watchRequest.Name,
		Namespace:                watchRequest.Namespace,
		Status:                   datamodel.Missing,
		GeneratedCompletionEvent: false,
	}
	byteKey, _ := json.Marshal(key)
	channel <- string(byteKey)
	channel <- progressionchecker.EndTransactionMarker
	json.NewEncoder(w).Encode(true)
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/deploy/{orgName}/{type}", deploy).Methods("POST")
	myRouter.HandleFunc("/delete/{group}/{version}/{kind}/{name}", deleteApplication).Methods("POST")
	myRouter.HandleFunc("/checkStatus/{group}/{version}/{kind}/{name}", checkStatus).Methods("GET")
	myRouter.HandleFunc("/watch/{orgName}", watch).Methods("POST")
	log.Fatal(http.ListenAndServe(":9091", myRouter))
}

func main() {
	kubernetesDriver := k8sdriver.New()
	drivers = Drivers{
		k8sDriver:  kubernetesDriver,
		argoDriver: argodriver.New(&kubernetesDriver),
	}
	channel = make(chan string)
	var getRestrictedKubernetesClient k8sdriver.KubernetesClientGetRestricted
	getRestrictedKubernetesClient = kubernetesDriver
	go progressionchecker.Start(getRestrictedKubernetesClient, channel)
	handleRequests()
}
