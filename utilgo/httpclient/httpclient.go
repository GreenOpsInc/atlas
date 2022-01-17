package httpclient

import (
	"crypto/tls"
	"log"
	"net/http"
	"time"

	"github.com/greenopsinc/util/tlsmanager"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type client struct {
	clientName tlsmanager.ClientName
	httpClient *http.Client
	tm         tlsmanager.Manager
}

func New(clientName tlsmanager.ClientName, tm tlsmanager.Manager) (HttpClient, error) {
	log.Printf("creating new client for %s", string(clientName))
	c := &client{clientName: clientName, tm: tm}
	httpClient, err := c.initHttpClient()
	if err != nil {
		return nil, err
	}
	c.httpClient = httpClient
	return c, err
}

func (c *client) Do(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}

func (c *client) initHttpClient() (*http.Client, error) {
	tlsConf, err := c.tm.GetClientTLSConf(c.clientName)
	if err != nil {
		return nil, err
	}

	httpClient := c.configureClient(tlsConf)
	if err = c.watchClient(); err != nil {
		return nil, err
	}
	return httpClient, nil
}

func (c *client) configureClient(tlsConf *tls.Config) *http.Client {
	log.Printf("configuring http client <%s>, tls insecure skip = %v", c.clientName, tlsConf.InsecureSkipVerify)
	return &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			TLSClientConfig: tlsConf,
		},
	}
}

func (c *client) watchClient() error {
	err := c.tm.WatchClientTLSConf(c.clientName, func(conf *tls.Config, err error) {
		log.Printf("in watchClient, conf = %v, err = %v\n", conf, err)
		if err != nil {
			log.Fatalf("an error occurred in the watch %s client: %s", c.clientName, err.Error())
		}
		c.httpClient = c.configureClient(conf)
	})
	return err
}
