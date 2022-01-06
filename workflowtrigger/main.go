package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/greenopsinc/util/starter"

	"github.com/gorilla/mux"
	"github.com/greenopsinc/util/db"
	"github.com/greenopsinc/util/kubernetesclient"
	"greenops.io/workflowtrigger/api"
	"greenops.io/workflowtrigger/api/argoauthenticator"
	"greenops.io/workflowtrigger/api/commanddelegator"
	"greenops.io/workflowtrigger/api/reposerver"
	"greenops.io/workflowtrigger/kafka"
	"greenops.io/workflowtrigger/schemavalidation"
	"greenops.io/workflowtrigger/tlsmanager"
)

func main() {
	var dbClient db.DbClient
	var kafkaClient kafka.KafkaClient
	var kubernetesClient kubernetesclient.KubernetesClient
	var repoManagerApi reposerver.RepoManagerApi
	var commandDelegatorApi commanddelegator.CommandDelegatorApi
	var argoAuthenticatorApi argoauthenticator.ArgoAuthenticatorApi
	var schemaValidator schemavalidation.RequestSchemaValidator
	var tlsManager tlsmanager.Manager
	dbClient = db.New(starter.GetDbClientConfig())
	kafkaClient = kafka.New(starter.GetKafkaClientConfig())
	kubernetesClient = kubernetesclient.New()
	argoAuthenticatorApi = argoauthenticator.New(tlsManager)
	repoManagerApi, err := reposerver.New(starter.GetRepoServerClientConfig(), tlsManager)
	if err != nil {
		log.Fatal(err)
	}
	commandDelegatorApi, err = commanddelegator.New(starter.GetCommandDelegatorServerClientConfig(), tlsManager)
	if err != nil {
		log.Fatal(err)
	}
	argoAuthenticatorApi = argoauthenticator.New(tlsManager)
	schemaValidator = schemavalidation.New(argoAuthenticatorApi, repoManagerApi)
	tlsManager, err = tlsmanager.New(kubernetesClient)
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.Use(argoAuthenticatorApi.Middleware)
	api.InitClients(dbClient, kafkaClient, kubernetesClient, repoManagerApi, commandDelegatorApi, schemaValidator)
	api.InitializeLocalCluster()
	api.InitPipelineTeamEndpoints(r)
	api.InitStatusEndpoints(r)
	api.InitClusterEndpoints(r)

	startAndWatchServer(tlsManager, r)
}

func startAndWatchServer(tlsManager tlsmanager.Manager, r *mux.Router) {
	log.Println("in startAndWatchServer")
	httpServerExitDone := &sync.WaitGroup{}
	httpServerExitDone.Add(1)

	tlsConfig, err := tlsManager.GetServerTLSConf()
	if err != nil {
		log.Fatal("failed to fetch tls configuration: ", err)
	}

	log.Println("in startAndWatchServer, before createServer")
	srv := createServer(tlsConfig, r)
	log.Println("in startAndWatchServer, before listenAndServe")
	go listenAndServe(httpServerExitDone, srv)

	tlsManager.WatchServerTLSConf(func(conf *tls.Config, err error) {
		log.Printf("in tlsManager.WatchServerTLSConf, conf = %v, err = %v\n", conf, err)
		if err != nil {
			defer httpServerExitDone.Done()
			log.Fatal(err)
		}

		httpServerExitDone.Add(1)

		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("in tlsManager.WatchServerTLSConf, srv.Shutdown. err = %v\n", err)
			defer httpServerExitDone.Done()
			log.Fatal(err)
		}
		log.Println("in tlsManager.WatchServerTLSConf, before createServer")
		srv := createServer(conf, r)
		log.Println("in tlsManager.WatchServerTLSConf, before listenAndServe")
		go listenAndServe(httpServerExitDone, srv)
	})

	httpServerExitDone.Wait()
}

// TODO: add logic to save cert to the atlas config folder
func createServer(tlsServerConf *tls.Config, handler http.Handler) *http.Server {
	return &http.Server{
		Handler:      handler,
		Addr:         ":8080",
		TLSConfig:    tlsServerConf,
		WriteTimeout: 20 * time.Second,
		ReadTimeout:  20 * time.Second,
	}
}

func listenAndServe(wg *sync.WaitGroup, srv *http.Server) {
	defer wg.Done()
	if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
