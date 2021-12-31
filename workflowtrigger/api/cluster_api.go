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
	"greenops.io/workflowtrigger/api/argoauthenticator"
)

const (
	localClusterName    string = "kubernetes_local"
	initLocalClusterEnv string = "ENABLE_LOCAL_CLUSTER"
)

func createCluster(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	var clusterSchema cluster.ClusterSchema
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	err = json.Unmarshal(buf.Bytes(), &clusterSchema)
	if err != nil || clusterSchema.NoDeploy != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !schemaValidator.VerifyRbac(argoauthenticator.CreateAction, argoauthenticator.ClusterResource, clusterSchema.ClusterName) {
		http.Error(w, "Not enough permissions", http.StatusForbidden)
		return
	}

	key := db.MakeDbClusterKey(orgName, clusterSchema.ClusterName)
	if dbClient.FetchClusterSchema(key).ClusterIP != "" {
		w.WriteHeader(http.StatusConflict)
		return
	}

	dbClient.StoreValue(key, clusterSchema)
	w.WriteHeader(http.StatusOK)
	return
}

func readCluster(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	clusterName := vars[clusterNameField]

	if !schemaValidator.VerifyRbac(argoauthenticator.GetAction, argoauthenticator.ClusterResource, clusterName) {
		http.Error(w, "Not enough permissions", http.StatusForbidden)
		return
	}

	key := db.MakeDbClusterKey(orgName, clusterName)
	clusterSchema := dbClient.FetchClusterSchema(key)
	emptyStruct := cluster.ClusterSchema{}
	if clusterSchema == emptyStruct {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clusterSchema)
}

func deleteCluster(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	clusterName := vars[clusterNameField]

	if !schemaValidator.VerifyRbac(argoauthenticator.DeleteAction, argoauthenticator.ClusterResource, clusterName) {
		http.Error(w, "Not enough permissions", http.StatusForbidden)
		return
	}

	key := db.MakeDbClusterKey(orgName, clusterName)
	dbClient.StoreValue(key, nil)
	w.WriteHeader(http.StatusOK)
}

func markNoDeployCluster(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	clusterName := vars[clusterNameField]
	var noDeployRequest cluster.NoDeployInfo
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	err = json.Unmarshal(buf.Bytes(), &noDeployRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//if !schemaValidator.VerifyRbac(argoauthenticator.UpdateAction, argoauthenticator.ClusterResource, clusterName) {
	//	http.Error(w, "Not enough permissions", http.StatusForbidden)
	//	return
	//}

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

	notification := getNotification(requestId)
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
	var noDeployRequest cluster.NoDeployInfo
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	err = json.Unmarshal(buf.Bytes(), &noDeployRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//if !schemaValidator.VerifyRbac(argoauthenticator.UpdateAction, argoauthenticator.ClusterResource, clusterName) {
	//	http.Error(w, "Not enough permissions", http.StatusForbidden)
	//	return
	//}

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

	notification := getNotification(requestId)
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
	if val := os.Getenv(initLocalClusterEnv); val == "" || val == "true" {
		key := db.MakeDbClusterKey("org", localClusterName)
		if dbClient.FetchClusterSchema(key).ClusterIP != "" {
			log.Printf("Local cluster already exists")
			return
		}
		dbClient.StoreValue(key, cluster.ClusterSchema{ClusterName: localClusterName})
	}
}

func InitClusterEndpoints(r *mux.Router) {
	r.HandleFunc("/cluster/{orgName}", createCluster).Methods("POST")
	r.HandleFunc("/cluster/{orgName}/{clusterName}", readCluster).Methods("GET")
	r.HandleFunc("/cluster/{orgName}/{clusterName}", deleteCluster).Methods("DELETE")
	r.HandleFunc("/cluster/{orgName}/{clusterName}/noDeploy/apply", markNoDeployCluster).Methods("POST")
	r.HandleFunc("/cluster/{orgName}/{clusterName}/noDeploy/remove", removeNoDeployCluster).Methods("POST")
}
