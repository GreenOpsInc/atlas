package main

import (
	"github.com/gorilla/mux"
	"github.com/greenopsinc/pipelinereposerver/api"
	"github.com/greenopsinc/util/db"
	"github.com/greenopsinc/util/httpserver"
	"github.com/greenopsinc/util/kubernetesclient"
	"github.com/greenopsinc/util/starter"
	"github.com/greenopsinc/util/tlsmanager"
	"log"
	"net/http"
)

func main() {
	var dbOperator db.DbOperator
	var tlsManager tlsmanager.Manager
	var kubernetesClient kubernetesclient.KubernetesClient
	if starter.GetNoAuthClientConfig() == "True" {
		kubernetesClient = nil
		tlsManager = tlsmanager.NoAuth()
	} else {
		kubernetesClient = kubernetesclient.New()
		tlsManager = tlsmanager.New(kubernetesClient)
	}
	dbOperator = db.New(starter.GetDbClientConfig())
	r := mux.NewRouter()
	api.InitClients(dbOperator, kubernetesClient)
	r.Use(Middleware)
	log.Println("setup middleware...")
	api.InitRepoEndpoints(r)
	api.InitFileEndpoints(r)

	httpserver.CreateAndWatchServer(tlsmanager.ClientRepoServer, tlsManager, r)
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				switch err.(type) {
				case string:
					log.Printf("Interal error occurred: %s", err.(string))
					http.Error(w, err.(string), http.StatusInternalServerError)
				case error:
					log.Printf("Interal error occurred: %s", err.(error).Error())
					http.Error(w, err.(error).Error(), http.StatusInternalServerError)
				default:
					http.Error(w, err.(string), http.StatusInternalServerError)
				}
			}
			return
		}()
		next.ServeHTTP(w, r)
	})
}
