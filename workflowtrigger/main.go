package main

import (
	"github.com/gorilla/mux"
	"greenops.io/workflowtrigger/api"
	"greenops.io/workflowtrigger/api/argoauthenticator"
	"greenops.io/workflowtrigger/api/reposerver"
	"greenops.io/workflowtrigger/db"
	"greenops.io/workflowtrigger/kafka"
	"greenops.io/workflowtrigger/kubernetesclient"
	"greenops.io/workflowtrigger/schemavalidation"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	var dbClient db.DbClient
	var kafkaClient kafka.KafkaClient
	var kubernetesClient kubernetesclient.KubernetesClient
	var repoManagerApi reposerver.RepoManagerApi
	var argoAuthenticatorApi argoauthenticator.ArgoAuthenticatorApi
	var schemaValidator schemavalidation.RequestSchemaValidator
	dbClient = db.New(GetDbClientConfig())
	kafkaClient = kafka.New(GetKafkaClientConfig())
	kubernetesClient = kubernetesclient.New()
	repoManagerApi = reposerver.New(GetRepoServerClientConfig())
	argoAuthenticatorApi = argoauthenticator.New()
	schemaValidator = schemavalidation.New(argoAuthenticatorApi, repoManagerApi)
	r := mux.NewRouter()
	r.Use(argoAuthenticatorApi.(*argoauthenticator.ArgoAuthenticatorApiImpl).Middleware)
	api.InitClients(dbClient, kafkaClient, kubernetesClient, repoManagerApi, schemaValidator)
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

const (
	//EnvVar Names
	dbAddress         string = "ATLAS_DB_ADDRESS"
	dbPassword        string = "ATLAS_DB_PASSWORD"
	kafkaAddress      string = "KAFKA_BOOTSTRAP_SERVERS"
	repoServerAddress string = "REPO_SERVER_ENDPOINT"

	//Default Names
	dbDefaultAddress         string = "localhost:6379"
	dbDefaultPassword        string = ""
	kafkaDefaultAddress      string = "localhost:29092"
	repoServerDefaultAddress string = "http://localhost:8081"
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
