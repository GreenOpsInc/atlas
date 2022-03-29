package api

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/greenopsinc/util/clientrequest"
	"github.com/greenopsinc/util/cluster"
	"github.com/greenopsinc/util/db"
	"greenops.io/workflowtrigger/api/argo"
)

const (
	initLocalClusterEnv string = "ENABLE_LOCAL_CLUSTER"
)

func createCluster(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	dbClient := dbOperator.GetClient()
	defer dbClient.Close()

	var clusterRequest cluster.ClusterCreateRequest
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	err = json.Unmarshal(buf.Bytes(), &clusterRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !schemaValidator.VerifyRbac(argo.CreateAction, argo.ClusterResource, clusterRequest.Name) {
		http.Error(w, "Not enough permissions", http.StatusForbidden)
		return
	}

	key := db.MakeDbClusterKey(orgName, clusterRequest.Name)
	if dbClient.FetchClusterSchema(key).ClusterIP != "" {
		http.Error(w, "cluster already exists", http.StatusConflict)
		return
	}

	if clusterRequest.Config != nil {
		//Creates the cluster in the the context of ArgoCD
		argoClusterApi.CreateCluster(clusterRequest.Name, clusterRequest.Server, *clusterRequest.Config)
	}
	//Creates the cluster in the context of Atlas
	clusterSchema := cluster.ClusterSchema{
		ClusterIP:   clusterRequest.Server,
		ClusterName: clusterRequest.Name,
		NoDeploy:    nil,
	}
	dbClient.StoreValue(key, clusterSchema)
	w.WriteHeader(http.StatusOK)
	return
}

func readCluster(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	clusterName := vars[clusterNameField]
	dbClient := dbOperator.GetClient()
	defer dbClient.Close()

	if !schemaValidator.VerifyRbac(argo.GetAction, argo.ClusterResource, clusterName) {
		http.Error(w, "Not enough permissions", http.StatusForbidden)
		return
	}

	key := db.MakeDbClusterKey(orgName, clusterName)
	clusterSchema := dbClient.FetchClusterSchema(key)
	emptyStruct := cluster.ClusterSchema{}
	if clusterSchema == emptyStruct {
		http.Error(w, "no cluster found", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clusterSchema)
}

func deleteCluster(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	clusterName := vars[clusterNameField]
	dbClient := dbOperator.GetClient()
	defer dbClient.Close()

	if !schemaValidator.VerifyRbac(argo.DeleteAction, argo.ClusterResource, clusterName) {
		http.Error(w, "Not enough permissions", http.StatusForbidden)
		return
	}

	key := db.MakeDbClusterKey(orgName, clusterName)
	clusterSchema := dbClient.FetchClusterSchema(key)
	argoClusterApi.DeleteCluster(clusterSchema.ClusterName, clusterSchema.ClusterIP)
	dbClient.StoreValue(key, nil)
	w.WriteHeader(http.StatusOK)
}

func markNoDeployCluster(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	clusterName := vars[clusterNameField]
	dbClient := dbOperator.GetClient()
	defer dbClient.Close()

	var noDeployRequest cluster.NoDeployInfo
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	err = json.Unmarshal(buf.Bytes(), &noDeployRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !schemaValidator.VerifyRbac(argo.UpdateAction, argo.ClusterResource, clusterName) {
		http.Error(w, "Not enough permissions", http.StatusForbidden)
		return
	}

	key := db.MakeDbClusterKey(orgName, clusterName)
	clusterSchema := dbClient.FetchClusterSchema(key)
	if clusterSchema.NoDeploy != nil && *(clusterSchema.NoDeploy) != noDeployRequest {
		http.Error(w, "Cluster already marked with no deploy", http.StatusBadRequest)
		return
	}
	clusterSchema.NoDeploy = &noDeployRequest
	dbClient.StoreValue(key, clusterSchema)

	requestId := commandDelegatorApi.SendNotification(orgName, clusterSchema.ClusterName, &clientrequest.ClientMarkNoDeployRequest{
		ClusterName: clusterName,
		Namespace:   noDeployRequest.Namespace,
		Apply:       true,
	})

	notification := getNotification(requestId, dbClient)
	if !notification.Successful {
		http.Error(w, notification.Body, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

func removeNoDeployCluster(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	clusterName := vars[clusterNameField]
	dbClient := dbOperator.GetClient()
	defer dbClient.Close()

	var noDeployRequest cluster.NoDeployInfo
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	err = json.Unmarshal(buf.Bytes(), &noDeployRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !schemaValidator.VerifyRbac(argo.UpdateAction, argo.ClusterResource, clusterName) {
		http.Error(w, "Not enough permissions", http.StatusForbidden)
		return
	}

	key := db.MakeDbClusterKey(orgName, clusterName)
	clusterSchema := dbClient.FetchClusterSchema(key)
	if clusterSchema.NoDeploy == nil {
		http.Error(w, "Cluster not marked with no deploy", http.StatusBadRequest)
		return
	}
	if *(clusterSchema.NoDeploy) != noDeployRequest {
		http.Error(w, "The no deploy payload has to match to ensure that you have read reason/name", http.StatusBadRequest)
		return
	}

	requestId := commandDelegatorApi.SendNotification(orgName, clusterSchema.ClusterName, &clientrequest.ClientMarkNoDeployRequest{
		ClusterName: clusterName,
		Namespace:   noDeployRequest.Namespace,
		Apply:       false,
	})

	notification := getNotification(requestId, dbClient)
	if !notification.Successful {
		http.Error(w, notification.Body, http.StatusInternalServerError)
		return
	}

	clusterSchema.NoDeploy = nil
	dbClient.StoreValue(key, clusterSchema)

	w.WriteHeader(http.StatusOK)
	return
}

func InitializeLocalCluster() {
	dbClient := dbOperator.GetClient()
	defer dbClient.Close()

	if val := os.Getenv(initLocalClusterEnv); val == "" || val == "true" {
		key := db.MakeDbClusterKey("org", cluster.LocalClusterName)
		if dbClient.FetchClusterSchema(key).ClusterIP != "" {
			log.Printf("Local cluster already exists")
			return
		}
		dbClient.StoreValue(key, cluster.ClusterSchema{ClusterName: cluster.LocalClusterName})
	}
}

func InitClusterEndpoints(r *mux.Router) {
	r.HandleFunc("/cluster/{orgName}", createCluster).Methods("POST", "OPTIONS")
	r.HandleFunc("/cluster/{orgName}/{clusterName}", readCluster).Methods("GET", "OPTIONS")
	r.HandleFunc("/cluster/{orgName}/{clusterName}", deleteCluster).Methods("DELETE", "OPTIONS")
	r.HandleFunc("/cluster/{orgName}/{clusterName}/noDeploy/apply", markNoDeployCluster).Methods("POST", "OPTIONS")
	r.HandleFunc("/cluster/{orgName}/{clusterName}/noDeploy/remove", removeNoDeployCluster).Methods("POST", "OPTIONS")
}
