package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"greenops.io/client/argodriver"
	"greenops.io/client/k8sdriver"
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

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/deploy/{group}/{version}/{kind}", deploy)
	log.Fatal(http.ListenAndServe(":9091", myRouter))
}

func main() {
	drivers = Drivers{
		k8sDriver:  k8sdriver.New(),
		argoDriver: argodriver.New(),
	}
	handleRequests()
}
