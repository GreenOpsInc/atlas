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
	"greenops.io/workflowtrigger/apikeysmanager"
	"greenops.io/workflowtrigger/schemavalidation"
)

func main() {
	kubernetesClient := kubernetesclient.New()
	tlsManager := tlsmanager.New(kubernetesClient)
	dbOperator := db.New(starter.GetDbClientConfig())
	kafkaClient, err := kafkaclient.New(starter.GetKafkaClientConfig(), tlsManager)
	if err != nil {
		log.Fatal(err)
	}
	repoManagerApi, err := reposerver.New(starter.GetRepoServerClientConfig(), tlsManager)
	if err != nil {
		log.Fatal(err)
	}
	apikeysManager := apikeysmanager.New(kubernetesClient)
	if err = apikeysManager.GenerateDefaultKeys(); err != nil {
		log.Fatal(err)
	}
	if err = apikeysManager.WatchApiKeys(); err != nil {
		log.Fatal(err)
	}
	commandDelegatorApi, err := commanddelegator.New(starter.GetCommandDelegatorServerClientConfig(), tlsManager, apikeysManager)
	if err != nil {
		log.Fatal(err)
	}
	argoAuthenticatorApi := argo.New(tlsManager, apikeysManager).GetAuthenticatorApi()
	schemaValidator := schemavalidation.New(argoAuthenticatorApi, repoManagerApi)
	r := mux.NewRouter()
	api.InitClients(dbOperator, kafkaClient, kubernetesClient, repoManagerApi, argo.New(tlsManager, apikeysManager).GetClusterApi(), commandDelegatorApi, schemaValidator)
	r.Use(argoAuthenticatorApi.(*argo.ArgoApiImpl).Middleware)
	api.InitializeLocalCluster()
	api.InitPipelineTeamEndpoints(r)
	api.InitStatusEndpoints(r)
	api.InitClusterEndpoints(r, apikeysManager)

	httpserver.CreateAndWatchServer(tlsmanager.ClientWorkflowTrigger, tlsManager, r)
}
