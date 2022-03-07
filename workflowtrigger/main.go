package main

import (
	"log"

	"github.com/gorilla/mux"
	"github.com/greenopsinc/util/db"
	"github.com/greenopsinc/util/httpserver"
	"github.com/greenopsinc/util/kafkaclient"
	"github.com/greenopsinc/util/kubernetesclient"
	"github.com/greenopsinc/util/starter"
	"github.com/greenopsinc/util/tlsmanager"
	"greenops.io/workflowtrigger/api"
	"greenops.io/workflowtrigger/api/argo"
	"greenops.io/workflowtrigger/api/commanddelegator"
	"greenops.io/workflowtrigger/api/reposerver"
	"greenops.io/workflowtrigger/schemavalidation"
)

func main() {
	var dbOperator db.DbOperator
	var kubernetesClient kubernetesclient.KubernetesClient
	var kafkaClient kafkaclient.KafkaClient
	var tlsManager tlsmanager.Manager
	var repoManagerApi reposerver.RepoManagerApi
	var commandDelegatorApi commanddelegator.CommandDelegatorApi
	var argoAuthenticatorApi argo.ArgoAuthenticatorApi
	var schemaValidator schemavalidation.RequestSchemaValidator

	kubernetesClient = kubernetesclient.New()
	if starter.GetNoAuthClientConfig() {
		tlsManager = tlsmanager.NoAuth()
	} else {
		tlsManager = tlsmanager.New(kubernetesClient)
	}
	dbOperator = db.New(starter.GetDbClientConfig())
	kafkaClient, err := kafkaclient.New(starter.GetKafkaClientConfig(), tlsManager)
	if err != nil {
		log.Fatal(err)
	}
	repoManagerApi, err = reposerver.New(starter.GetRepoServerClientConfig(), tlsManager)
	if err != nil {
		log.Fatal(err)
	}
	commandDelegatorApi, err = commanddelegator.New(starter.GetCommandDelegatorServerClientConfig(), tlsManager)
	if err != nil {
		log.Fatal(err)
	}
	coreArgoClient := argo.New(tlsManager)
	argoAuthenticatorApi = coreArgoClient.GetAuthenticatorApi()
	schemaValidator = schemavalidation.New(argoAuthenticatorApi, repoManagerApi)
	r := mux.NewRouter()
	api.InitClients(dbOperator, kafkaClient, kubernetesClient, repoManagerApi, coreArgoClient.GetClusterApi(), coreArgoClient.GetRepoApi(), commandDelegatorApi, schemaValidator)
	r.Use(argoAuthenticatorApi.(*argo.ArgoApiImpl).Middleware)
	log.Println("setup middleware...")
	api.InitializeLocalCluster()
	api.InitPipelineTeamEndpoints(r)
	api.InitStatusEndpoints(r)
	api.InitClusterEndpoints(r)

	httpserver.CreateAndWatchServer(tlsmanager.ClientWorkflowTrigger, tlsManager, r)
}
