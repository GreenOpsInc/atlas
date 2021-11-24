package api

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	"greenops.io/workflowtrigger/db"
	"greenops.io/workflowtrigger/util/cluster"
	"net/http"
)

const (
	clusterNameField string = "clusterName"
)

func createCluster(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	var clusterSchema cluster.ClusterSchema
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	err = json.Unmarshal(buf.Bytes(), &clusterSchema)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
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

	key := db.MakeDbClusterKey(orgName, clusterName)
	clusterSchema := dbClient.FetchClusterSchema(key)
	if clusterSchema.ClusterIP == "" {
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

	key := db.MakeDbClusterKey(orgName, clusterName)
	dbClient.StoreValue(key, nil)
	w.WriteHeader(http.StatusOK)
}

func InitClusterEndpoints(r *mux.Router) {
	r.HandleFunc("/cluster/{orgName}", createCluster).Methods("POST")
	r.HandleFunc("/cluster/{orgName}/{clusterName}", readCluster).Methods("GET")
	r.HandleFunc("/cluster/{orgName}/{clusterName}", deleteCluster).Methods("DELETE")
}
