package main

import (
	"bytes"
	"net/http"

	"github.com/greenopsinc/util/kubernetesclient"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/greenopsinc/util/clientrequest"
	"github.com/greenopsinc/util/cluster"
	"github.com/greenopsinc/util/db"
	"github.com/greenopsinc/util/httpserver"
	"github.com/greenopsinc/util/serializer"
	"github.com/greenopsinc/util/serializerutil"
	"github.com/greenopsinc/util/starter"
	"github.com/greenopsinc/util/tlsmanager"
)

const (
	orgNameField     string = "orgName"
	clusterNameField string = "clusterName"
)

var dbOperator db.DbOperator
var emptyClusterStruct = cluster.ClusterSchema{}
var emptyClientRequestPacketStruct = clientrequest.ClientRequestPacket{}

func getCommands(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	clusterName := vars[clusterNameField]
	dbClient := dbOperator.GetClient()
	defer dbClient.Close()

	clusterKey := db.MakeDbClusterKey(orgName, clusterName)
	clusterSchema := dbClient.FetchClusterSchemaTransactionless(clusterKey)
	if clusterSchema == emptyClusterStruct {
		http.Error(w, "Cluster was nil", http.StatusBadRequest)
		return
	}
	noDeployInfo := clusterSchema.NoDeploy
	notificationQueueKey := db.MakeClientNotificationQueueKey(orgName, clusterName)
	requestQueueKey := db.MakeClientRequestQueueKey(orgName, clusterName)

	//Notifications are from the management plane, they take priority over basic processing
	//Notifications also do not change the state of deployments and clusters, so can be processed even when in a no-deploy state
	clientNotificationHead := dbClient.FetchHeadInClientRequestList(notificationQueueKey)
	if clientNotificationHead != emptyClientRequestPacketStruct {
		//Notifications don't subscribe to the retry/final try logic, the client wrapper makes sure it's only run once
		writeResponse(w, []clientrequest.ClientRequestEvent{clientNotificationHead.ClientRequest})
		return
	}
	if noDeployInfo != nil && noDeployInfo.Namespace == "" {
		writeResponse(w, nil)
		return
	}
	clientRequestPacket := dbClient.FetchHeadInClientRequestList(requestQueueKey)
	if clientRequestPacket != emptyClientRequestPacketStruct {
		if noDeployInfo != nil && clientRequestPacket.Namespace == noDeployInfo.Namespace {
			writeResponse(w, nil)
			return
		}
		request := clientRequestPacket.ClientRequest
		request.SetFinalTry(clientRequestPacket.RetryCount >= 5)
		writeResponse(w, []clientrequest.ClientRequestEvent{request})
		return
	}
}

func ackHeadOfRequestList(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	clusterName := vars[clusterNameField]
	dbClient := dbOperator.GetClient()
	defer dbClient.Close()

	clusterKey := db.MakeDbClusterKey(orgName, clusterName)
	clusterSchema := dbClient.FetchClusterSchemaTransactionless(clusterKey)
	if clusterSchema == emptyClusterStruct {
		http.Error(w, "Cluster was nil", http.StatusBadRequest)
		return
	}
	key := db.MakeClientRequestQueueKey(orgName, clusterName)
	dbClient.UpdateHeadInTransactionlessList(key, nil)
}

func addNotificationCommand(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	clusterName := vars[clusterNameField]
	dbClient := dbOperator.GetClient()
	defer dbClient.Close()

	clusterKey := db.MakeDbClusterKey(orgName, clusterName)
	clusterSchema := dbClient.FetchClusterSchemaTransactionless(clusterKey)
	if clusterSchema == emptyClusterStruct {
		http.Error(w, "Cluster was nil", http.StatusBadRequest)
		return
	}
	key := db.MakeClientNotificationQueueKey(orgName, clusterName)
	requestId := uuid.New().String()

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, "Error when reading request body", http.StatusBadRequest)
		return
	}
	notification := serializer.Deserialize(buf.String(), serializerutil.ClientRequestType).(clientrequest.ClientRequestEvent)
	notification.(clientrequest.NotificationRequestEvent).SetRequestId(requestId)
	dbClient.InsertValueInTransactionlessList(key, clientrequest.ClientRequestPacket{
		Namespace:     "",
		ClientRequest: notification,
	})

	writeResponse(w, requestId)
}

func ackHeadOfNotificationList(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	clusterName := vars[clusterNameField]
	dbClient := dbOperator.GetClient()
	defer dbClient.Close()

	clusterKey := db.MakeDbClusterKey(orgName, clusterName)
	clusterSchema := dbClient.FetchClusterSchemaTransactionless(clusterKey)
	if clusterSchema == emptyClusterStruct {
		http.Error(w, "Cluster was nil", http.StatusBadRequest)
		return
	}
	key := db.MakeClientNotificationQueueKey(orgName, clusterName)
	dbClient.UpdateHeadInTransactionlessList(key, nil)
}

func retryMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	clusterName := vars[clusterNameField]
	dbClient := dbOperator.GetClient()
	defer dbClient.Close()

	clusterKey := db.MakeDbClusterKey(orgName, clusterName)
	clusterSchema := dbClient.FetchClusterSchemaTransactionless(clusterKey)
	if clusterSchema == emptyClusterStruct {
		http.Error(w, "Cluster was nil", http.StatusBadRequest)
		return
	}
	key := db.MakeClientRequestQueueKey(orgName, clusterName)
	clientRequestPacket := dbClient.FetchHeadInClientRequestList(key)
	if clientRequestPacket != emptyClientRequestPacketStruct {
		clientRequestPacket.RetryCount = clientRequestPacket.RetryCount + 1
	}
	dbClient.InsertValueInTransactionlessList(key, clientRequestPacket)
	dbClient.UpdateHeadInTransactionlessList(key, nil)
}

func writeResponse(w http.ResponseWriter, i interface{}) {
	if i != nil {
		payload := serializer.Serialize(i)
		w.Write([]byte(payload))
	}
	w.Header().Set("Content-Type", "application/json")
}

func InitEndpoints(r *mux.Router) {
	r.HandleFunc("/requests/{orgName}/{clusterName}", getCommands).Methods("GET")
	r.HandleFunc("/requests/ackHead/{orgName}/{clusterName}", ackHeadOfRequestList).Methods("DELETE")
	r.HandleFunc("/notifications/{orgName}/{clusterName}", addNotificationCommand).Methods("POST")
	r.HandleFunc("/notifications/ackHead/{orgName}/{clusterName}", ackHeadOfNotificationList).Methods("DELETE")
	r.HandleFunc("/requests/retry/{orgName}/{clusterName}", retryMessage).Methods("DELETE")
}

func main() {
	dbOperator = db.New(starter.GetDbClientConfig())
	var tlsManager tlsmanager.Manager
	var kubernetesClient kubernetesclient.KubernetesClient
	kubernetesClient = kubernetesclient.New()

	if starter.GetNoAuthClientConfig() {
		tlsManager = tlsmanager.NoAuth()
	} else {
		tlsManager = tlsmanager.New(kubernetesClient)
	}
	r := mux.NewRouter()
	InitEndpoints(r)
	httpserver.CreateAndWatchServer(tlsmanager.ClientCommandDelegator, tlsManager, r)
}
