package util

import (
	"crypto/tls"
	"net/http"
	"time"

	"greenops.io/client/tlsmanager"
)

func CheckFatalError(err error) {
	if err != nil {
		panic(err)
	}
}

func CreateHttpClient(tm tlsmanager.Manager) *http.Client {
	cert := tm.GetCertificatePEM()
	certpool := tm.BestEffortSystemCertPool()
	certpool.AppendCertsFromPEM(cert)
	clientTLSConf := &tls.Config{
		RootCAs: certpool,
	}
	transport := &http.Transport{
		TLSClientConfig: clientTLSConf,
	}
	return &http.Client{
		Timeout:   time.Second * 10,
		Transport: transport,
	}
}
