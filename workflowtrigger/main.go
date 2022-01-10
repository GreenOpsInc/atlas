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
	"greenops.io/workflowtrigger/api/argoauthenticator"
	"greenops.io/workflowtrigger/api/commanddelegator"
	"greenops.io/workflowtrigger/api/reposerver"
	"greenops.io/workflowtrigger/schemavalidation"
)

func main() {
	var dbClient db.DbClient
	var kafkaClient kafkaclient.KafkaClient
	var kubernetesClient kubernetesclient.KubernetesClient
	var repoManagerApi reposerver.RepoManagerApi
	var commandDelegatorApi commanddelegator.CommandDelegatorApi
	var argoAuthenticatorApi argoauthenticator.ArgoAuthenticatorApi
	var schemaValidator schemavalidation.RequestSchemaValidator
	var tlsManager tlsmanager.Manager
	kubernetesClient = kubernetesclient.New()
	dbClient = db.New(starter.GetDbClientConfig())
	tlsManager = tlsmanager.New(kubernetesClient)
	kafkaClient, err := kafkaclient.New(starter.GetKafkaClientConfig(), tlsManager)
	if err != nil {
		log.Fatal(err)
	}
	argoAuthenticatorApi = argoauthenticator.New(tlsManager)
	repoManagerApi, err = reposerver.New(starter.GetRepoServerClientConfig(), tlsManager)
	if err != nil {
		log.Fatal(err)
	}
	commandDelegatorApi, err = commanddelegator.New(starter.GetCommandDelegatorServerClientConfig(), tlsManager)
	if err != nil {
		log.Fatal(err)
	}
	argoAuthenticatorApi = argoauthenticator.New(tlsManager)
	schemaValidator = schemavalidation.New(argoAuthenticatorApi, repoManagerApi)
	r := mux.NewRouter()
	r.Use(argoAuthenticatorApi.Middleware)
	api.InitClients(dbClient, kafkaClient, kubernetesClient, repoManagerApi, commandDelegatorApi, schemaValidator)
	api.InitializeLocalCluster()
	api.InitPipelineTeamEndpoints(r)
	api.InitStatusEndpoints(r)
	api.InitClusterEndpoints(r)

	httpserver.CreateAndWatchServer(tlsmanager.ClientWorkflowTrigger, tlsManager, r)
}
