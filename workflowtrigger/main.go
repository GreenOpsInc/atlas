package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"greenops.io/workflowtrigger/api"
	"greenops.io/workflowtrigger/api/argoauthenticator"
	"greenops.io/workflowtrigger/api/commanddelegator"
	"greenops.io/workflowtrigger/api/reposerver"
	"greenops.io/workflowtrigger/db"
	"greenops.io/workflowtrigger/kafka"
	"greenops.io/workflowtrigger/kubernetesclient"
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
	dbClient = db.New(GetDbClientConfig())
	kafkaClient = kafka.New(GetKafkaClientConfig())
	kubernetesClient = kubernetesclient.New()
	argoAuthenticatorApi = argoauthenticator.New(tlsManager)
	schemaValidator = schemavalidation.New(argoAuthenticatorApi, repoManagerApi)
	repoManagerApi, err := reposerver.New(GetRepoServerClientConfig(), tlsManager)
	if err != nil {
		log.Fatal(err)
	}
	commandDelegatorApi, err = commanddelegator.New(GetCommandDelegatorServerClientConfig(), tlsManager)
	if err != nil {
		log.Fatal(err)
	}
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

const (
	//EnvVar Names
	dbAddress                     string = "ATLAS_DB_ADDRESS"
	dbPassword                    string = "ATLAS_DB_PASSWORD"
	kafkaAddress                  string = "KAFKA_BOOTSTRAP_SERVERS"
	repoServerAddress             string = "REPO_SERVER_ENDPOINT"
	commandDelegatorServerAddress string = "COMMAND_DELEGATOR_SERVER_ENDPOINT"
	configPath                    string = "ATLAS_CONFIG_PATH"

	//Default Names
	dbDefaultAddress                     string = "localhost:6379"
	dbDefaultPassword                    string = ""
	kafkaDefaultAddress                  string = "localhost:29092"
	repoServerDefaultAddress             string = "http://localhost:8081"
	commandDelegatorServerDefaultAddress string = "http://localhost:8080"
	configDefaultPath                    string = "/home/.atlas"
)

func GetDbClientConfig() (string, string) {
	address := dbDefaultAddress
	password := dbDefaultPassword
	if val := os.Getenv(dbAddress); val != "" {
		address = val
	}
	if val := os.Getenv(dbPassword); val != "" {
		password = val
	}
	return address, password
}

func GetKafkaClientConfig() string {
	address := kafkaDefaultAddress
	if val := os.Getenv(kafkaAddress); val != "" {
		address = val
	}
	return address
}

func GetRepoServerClientConfig() string {
	address := repoServerDefaultAddress
	if val := os.Getenv(repoServerAddress); val != "" {
		address = val
	}
	return address
}

func GetCommandDelegatorServerClientConfig() string {
	address := commandDelegatorServerDefaultAddress
	if val := os.Getenv(commandDelegatorServerAddress); val != "" {
		address = val
	}
	return address
}
