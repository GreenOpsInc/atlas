package main

import (
	"github.com/greenopsinc/util/starter"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/greenopsinc/util/db"
	"github.com/greenopsinc/util/kubernetesclient"
	"greenops.io/workflowtrigger/api"
	"greenops.io/workflowtrigger/api/argo"
	"greenops.io/workflowtrigger/api/commanddelegator"
	"greenops.io/workflowtrigger/api/reposerver"
	"greenops.io/workflowtrigger/kafka"
	"greenops.io/workflowtrigger/schemavalidation"
)

func main() {
	var dbOperator db.DbOperator
	var kafkaClient kafka.KafkaClient
	var kubernetesClient kubernetesclient.KubernetesClient
	var repoManagerApi reposerver.RepoManagerApi
	var commandDelegatorApi commanddelegator.CommandDelegatorApi
	var argoAuthenticatorApi argo.ArgoAuthenticatorApi
	var schemaValidator schemavalidation.RequestSchemaValidator
	dbOperator = db.New(starter.GetDbClientConfig())
	kafkaClient = kafka.New(starter.GetKafkaClientConfig())
	kubernetesClient = kubernetesclient.New()
	repoManagerApi = reposerver.New(starter.GetRepoServerClientConfig())
	commandDelegatorApi = commanddelegator.New(starter.GetCommandDelegatorServerClientConfig())
	argoAuthenticatorApi = argo.New().GetAuthenticatorApi()
	schemaValidator = schemavalidation.New(argoAuthenticatorApi, repoManagerApi)
	r := mux.NewRouter()
	r.Use(argoAuthenticatorApi.(*argo.ArgoApiImpl).Middleware)
	api.InitClients(dbOperator, kafkaClient, kubernetesClient, repoManagerApi, argo.New().GetClusterApi(), commandDelegatorApi, schemaValidator)
	api.InitializeLocalCluster()
	api.InitPipelineTeamEndpoints(r)
	api.InitStatusEndpoints(r)
	api.InitClusterEndpoints(r)

	srv := &http.Server{
		Handler:      r,
		Addr:         ":8080",
		WriteTimeout: 20 * time.Second,
		ReadTimeout:  20 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
