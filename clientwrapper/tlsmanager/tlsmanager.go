package tlsmanager

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"log"
	"time"

	kclient "greenops.io/client/kubernetesclient"
	corev1 "k8s.io/api/core/v1"
)

// TODO: add certPEM conf for each api client in workflowtrigger to perform secure requests
type Manager interface {
	GetTLSConf() (*tls.Config, error)
	GetCertificatePEM() []byte
	WatchTLSConf(handler func(conf *tls.Config, err error))
	BestEffortSystemCertPool() *x509.CertPool
}

type tlsManager struct {
	k       kclient.KubernetesClient
	conf    *tls.Config
	certPEM []byte
}

// TODO: currently those values are hardcoded, fetch them from config or somewhere else
const (
	AtlasCustomTLSSecretName = "atlas-tls"
	AtlasNamespace           = "default"
)

func New(k kclient.KubernetesClient) Manager {
	return &tlsManager{k: k}
}

// TODO: currently we are fetching a conf from current cluster
//		but client could be deployed in the other cluster
//		in this case we'll need to call workflowtrigger api to get tls keypair
func (m *tlsManager) GetTLSConf() (*tls.Config, error) {
	log.Println("in GetTLSConf")
	if m.conf != nil {
		return m.conf, nil
	}
	return m.fetchTLSConf()
}

func (m *tlsManager) fetchTLSConf() (*tls.Config, error) {
	successCh := make(chan *tls.Config, 1)
	errCh := make(chan error, 1)
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	go func() {
		for {
			conf, certPEM, err := m.getTLSConfFromSecrets()
			log.Printf("in GetTLSConf, conf = %v\n", conf)
			if err != nil {
				errCh <- err
				break
			}
			if conf != nil {
				log.Println("CERT FOUND IN SECRETS")
				m.setTLSConf(conf, certPEM)
				successCh <- conf
				break
			}

			select {
			case <-time.After(time.Second * 2):
				continue
			case <-timeoutCtx.Done():
				break
			}
		}
	}()

	select {
	case res := <-successCh:
		return res, nil
	case err := <-errCh:
		return nil, err
	case <-timeoutCtx.Done():
		return nil, errors.New("tls certificate fetch failed: timeout reached")
	}
}

func (m *tlsManager) GetCertificatePEM() []byte {
	return m.certPEM
}

func (m *tlsManager) setTLSConf(conf *tls.Config, cert []byte) {
	m.certPEM = cert
	m.conf = conf
}

// TODO: to add a tls certPEM\key to secrets run:
//		kubectl create secret tls atlas-server-tls --certPEM ./certPEM.pem --key ./key.pem
// TODO: check that this watcher is not trigger server reloading if certPEM is not changed
//		it could possibly happend on server start when certPEM is available and we also receiving secret change event
// TODO: call this function only if we are in the same cluster as wokrflowtigger
func (m *tlsManager) WatchTLSConf(handler func(conf *tls.Config, err error)) {
	log.Println("in WatchTLSConf")
	m.k.WatchSecretData(AtlasCustomTLSSecretName, AtlasNamespace, func(secret *corev1.Secret) {
		log.Printf("in WatchTLSConf. data = %s\n", secret)
		log.Printf("in WatchTLSConf, secret data = %v\n", secret.Data)
		config, err := m.generateTLSConfFromKeyPair(secret.Data[kclient.TLSSecretCrtName], secret.Data[kclient.TLSSecretCrtName])
		log.Printf("in WatchTLSConf, conf = %v\n", config)
		if err != nil {
			handler(nil, err)
			return
		}
		m.setTLSConf(config, secret.Data[kclient.TLSSecretCrtName])
		handler(config, nil)
	})
}

func (m *tlsManager) getTLSConfFromSecrets() (*tls.Config, []byte, error) {
	log.Println("in getTLSConfFromSecrets")
	secret := m.k.FetchSecretData(AtlasCustomTLSSecretName, AtlasNamespace)
	log.Println("in getTLSConfFromSecrets, secret: ", secret)
	if secret == nil {
		return nil, nil, nil
	}

	conf, err := m.generateTLSConfFromKeyPair(secret[kclient.TLSSecretCrtName], secret[kclient.TLSSecretKeyName])
	if err != nil {
		return nil, nil, err
	}

	m.k.StoreServerTLSConf(string(secret[kclient.TLSSecretCrtName]), string(secret[kclient.TLSSecretKeyName]), AtlasNamespace)
	return conf, secret[kclient.TLSSecretCrtName], nil
}

func (m *tlsManager) generateTLSConfFromKeyPair(cert []byte, key []byte) (*tls.Config, error) {
	log.Printf("in generateTLSConfFromKeyPair certPEM = %s, key = %s\n", string(cert), string(key))
	c, err := tls.X509KeyPair(cert, key)
	log.Printf("in generateTLSConfFromKeyPair c = %v\n", c)
	if err != nil {
		return nil, err
	}

	rootCAs := m.BestEffortSystemCertPool()
	return &tls.Config{
		Certificates:             []tls.Certificate{c},
		MinVersion:               tls.VersionTLS13,
		PreferServerCipherSuites: true,
		RootCAs:                  rootCAs,
	}, nil
}

func (m *tlsManager) BestEffortSystemCertPool() *x509.CertPool {
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		log.Println("root ca not found, returning new...")
		return x509.NewCertPool()
	}
	log.Println("root ca found")
	return rootCAs
}
