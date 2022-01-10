package httpserver

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/greenopsinc/util/tlsmanager"
)

func CreateAndWatchServer(serverName tlsmanager.ClientName, tlsManager tlsmanager.Manager, r *mux.Router) {
	log.Println("in startAndWatchServer")
	httpServerExitDone := &sync.WaitGroup{}
	httpServerExitDone.Add(1)

	tlsConfig, err := tlsManager.GetServerTLSConf(serverName)
	if err != nil {
		log.Fatal("failed to fetch tls configuration: ", err)
	}

	log.Println("in startAndWatchServer, before createServer")
	srv := createServer(tlsConfig, r)
	log.Println("in startAndWatchServer, before listenAndServe")
	go listenAndServe(httpServerExitDone, srv)

	err = tlsManager.WatchServerTLSConf(serverName, func(conf *tls.Config, err error) {
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
	if err != nil {
		log.Fatal("failed to watch server tls configuration: ", err)
	}

	httpServerExitDone.Wait()
}

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
	log.Println("listening the server: ", srv)
	if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
