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
	//	pemS := `
	//-----BEGIN CERTIFICATE-----
	//MIICDDCCAXWgAwIBAgIJANpKauCmIVCHMA0GCSqGSIb3DQEBDQUAMDoxDjAMBgNV
	//BAMMBUF0bGFzMQ4wDAYDVQQKDAVBdGxhczELMAkGA1UECAwCU0YxCzAJBgNVBAYT
	//AlVTMCAXDTIxMDExMTE5MzQwN1oYDzk5OTkxMjMxMjM1OTU5WjA6MQ4wDAYDVQQD
	//DAVBdGxhczEOMAwGA1UECgwFQXRsYXMxCzAJBgNVBAgMAlNGMQswCQYDVQQGEwJV
	//UzCBnzANBgkqhkiG9w0BAQEFAAOBjQAwgYkCgYEAnLJHeND5HKF0xkJqsia0cN4b
	//4OSD2/usqsYX1LwmUvhZgXS7aU54U+g3ccTbslExxlOBE/8K+zIPkQyKF+/oLYaJ
	//XfVf7aUCCB9UNJGX9W988AHrd3m8vEfOnbKS8DkDl7LpGT5BpzHvlpB8q2ixbOeN
	//nH2Onvj9cupR/F9La7sCAwEAAaMYMBYwFAYDVR0RBA0wC4IJbG9jYWxob3N0MA0G
	//CSqGSIb3DQEBDQUAA4GBAJjOnoiYNDMDpFeDxlDRwwLV0BHj3iEM2ezjMwaMPS2j
	//CnhPutWgycl0i5NpTmaN9BGZwf4yybXQhjd6yyot/iE5Ll2iSNlMkKL2ZwcdmWBa
	//fXdS53qyb7CYu5yJpP2p+864lGViU4oLFcpqWKkOgl/fm/Rw8ssHzR9QJNXEv8P8
	//-----END CERTIFICATE-----`
	//
	//	block, _ := pem.Decode([]byte(pemS))
	//	cert, _ := x509.ParseCertificate(block.Bytes)
	//	tlsConf := &tls.Config{
	//		Certificates: []tls.Certificate{
	//			{
	//				Certificate: [][]byte{cert.Raw},
	//			},
	//		},
	//	}

	log.Printf("configuring http client <%s>, tls insecure skip = %v, root ca = %v", c.clientName, tlsConf.InsecureSkipVerify, tlsConf.RootCAs)
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
