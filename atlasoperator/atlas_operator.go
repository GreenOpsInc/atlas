package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"greenops.io/client/argodriver"
	"greenops.io/client/k8sdriver"
	"greenops.io/client/progressionchecker"
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
	groupVersionKind := schema.GroupVersionKind{
		Group:   vars["group"],
		Version: vars["version"],
		Kind:    vars["kind"],
	}
	byteReqBody, _ := ioutil.ReadAll(r.Body)
	stringReqBody := string(byteReqBody)
	var success bool
	if strings.Contains(groupVersionKind.Group, "argo") {
		success = drivers.argoDriver.Deploy(stringReqBody)
	} else {
		success = drivers.k8sDriver.Deploy(stringReqBody)
	}

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

func watchApplication(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channel <- vars["name"]
	channel <- progressionchecker.EndTransactionMarker
	json.NewEncoder(w).Encode(true)
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/deploy/{group}/{version}/{kind}", deploy).Methods("POST")
	myRouter.HandleFunc("/delete/{group}/{version}/{kind}/{name}", deleteApplication).Methods("POST")
	myRouter.HandleFunc("/checkStatus/{group}/{version}/{kind}/{name}", checkStatus).Methods("GET")
	myRouter.HandleFunc("/watchApplication/{group}/{version}/{kind}/{name}", watchApplication).Methods("POST")
	log.Fatal(http.ListenAndServe(":9091", myRouter))
}

func main() {
	drivers = Drivers{
		k8sDriver:  k8sdriver.New(),
		argoDriver: argodriver.New(),
	}
	channel = make(chan string)
	go progressionchecker.Start(channel)
	handleRequests()
}
