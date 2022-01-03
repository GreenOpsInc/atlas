package tlsmanager

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"time"

	kclient "greenops.io/workflowtrigger/kubernetesclient"
	corev1 "k8s.io/api/core/v1"
)

/* TODO

1. create a function to generate TLS certs for servers
2. create a function to get certificate from secrets
3. use secrets crt or self signed
4. create a function to update servers on certificate change
	use kubernetesclient/WatchSecretData function
*/

// TODO: add cert conf for each api client in workflowtrigger to perform secure requests
type Manager interface {
	GetTLSConf() (*tls.Config, error)
	WatchTLSConf(handler func(conf *tls.Config, err error))
}

type tlsManager struct {
	k              kclient.KubernetesClient
	conf           *tls.Config
	selfSignedConf *tls.Config
}

// TODO: currently those values are hardcoded, fetch them from config or somewhere else
const (
	AtlasCustomTLSSecretName = "atlas-tls"
	AtlasNamespace           = "default"
)

func New(k kclient.KubernetesClient) Manager {
	return &tlsManager{k: k}
}

func (m *tlsManager) GetTLSConf() (*tls.Config, error) {
	log.Println("in GetTLSConf")
	if m.conf != nil {
		return m.conf, nil
	}

	conf, err := m.getTLSConfFromSecrets()
	log.Printf("in GetTLSConf, conf = %v\n", conf)
	if err != nil {
		return nil, err
	}
	if conf != nil {
		log.Println("CERT FOUND IN SECRETS")
		m.setMainTLSConf(conf)
		return conf, nil
	}

	log.Println("in GetTLSConf, before getSelfSignedTLSConf")
	conf, err = m.getSelfSignedTLSConf()
	if err != nil {
		return nil, err
	}

	m.setMainTLSConf(conf)
	return conf, nil
}

func (m *tlsManager) setMainTLSConf(conf *tls.Config) {
	m.conf = conf
}

func (m *tlsManager) setSelfSignedTLSConf(conf *tls.Config) {
	m.selfSignedConf = conf
}

// TODO: to add a tls cert\key to secrets run:
//		kubectl create secret tls atlas-server-tls --cert ./cert.pem --key ./key.pem
// TODO: check that this watcher is not trigger server reloading if cert is not changed
//		it could possibly happend on server start when cert is available and we also receiving secret change event
func (m *tlsManager) WatchTLSConf(handler func(conf *tls.Config, err error)) {
	log.Println("in WatchTLSConf")
	m.k.WatchSecretData(AtlasCustomTLSSecretName, AtlasNamespace, func(t kclient.SecretChangeType, secret *corev1.Secret) {
		log.Printf("in WatchTLSConf, event %v. data = %s\n", t, secret)
		var config *tls.Config
		var err error

		switch t {
		case kclient.SecretChangeTypeAdd:
			fallthrough
		case kclient.SecretChangeTypeUpdate:
			log.Printf("in WatchTLSConf, secret data = %v\n", secret.Data)
			config, err = m.generateTLSConfFromKeyPair(secret.Data[kclient.TLSSecretCrtName], secret.Data[kclient.TLSSecretCrtName])
			log.Printf("in WatchTLSConf, conf = %v\n", config)
		case kclient.SecretChangeTypeDelete:
			config, err = m.getSelfSignedTLSConf()
		}

		if err != nil {
			handler(nil, err)
			return
		}
		m.setMainTLSConf(config)
		handler(config, nil)
	})
}

func (m *tlsManager) getSelfSignedTLSConf() (*tls.Config, error) {
	log.Println("in getSelfSignedTLSConf")
	if m.selfSignedConf != nil {
		return m.selfSignedConf, nil
	}

	conf, err := m.generateSelfSignedTLSConf()
	if err != nil {
		return nil, err
	}
	log.Printf("in getSelfSignedTLSConf, conf = %v\n", conf)

	m.setSelfSignedTLSConf(conf)
	return conf, nil
}

func (m *tlsManager) getTLSConfFromSecrets() (*tls.Config, error) {
	log.Println("in getTLSConfFromSecrets")
	secret := m.k.FetchSecretData(AtlasCustomTLSSecretName, AtlasNamespace)
	log.Println("in getTLSConfFromSecrets, secret: ", secret)
	if secret == nil {
		return nil, nil
	}

	conf, err := m.generateTLSConfFromKeyPair(secret[kclient.TLSSecretCrtName], secret[kclient.TLSSecretKeyName])
	if err != nil {
		return nil, err
	}

	m.k.StoreServerTLSConf(string(secret[kclient.TLSSecretCrtName]), string(secret[kclient.TLSSecretKeyName]), AtlasNamespace)
	return conf, nil
}

func (m *tlsManager) generateTLSConfFromKeyPair(cert []byte, key []byte) (*tls.Config, error) {
	log.Printf("in generateTLSConfFromKeyPair cert = %s, key = %s\n", string(cert), string(key))
	c, err := tls.X509KeyPair(cert, key)
	log.Printf("in generateTLSConfFromKeyPair c = %v\n", c)
	if err != nil {
		return nil, err
	}

	rootCAs := bestEffortSystemCertPool()
	return &tls.Config{
		Certificates:             []tls.Certificate{c},
		MinVersion:               tls.VersionTLS13,
		PreferServerCipherSuites: true,
		RootCAs:                  rootCAs,
	}, nil
}

func bestEffortSystemCertPool() *x509.CertPool {
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		log.Println("root ca not found, returning new...")
		return x509.NewCertPool()
	}
	log.Println("root ca found")
	return rootCAs
}

// TODO: add certificate to the global registry or create a new registry
// TODO: try to use most secure configuration for ca and certificate conf
func (m *tlsManager) generateSelfSignedTLSConf() (*tls.Config, error) {
	certSerialNumber, err := generateCertificateSerialNumber()
	if err != nil {
		return nil, err
	}

	cert := &x509.Certificate{
		SerialNumber: certSerialNumber,
		Subject: pkix.Name{
			Organization: []string{"GreenOps, INC."},
			Country:      []string{"US"},
		},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
	}
	log.Printf("in generateSelfSignedTLSConf, cert = %v\n", cert)

	certPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}
	log.Printf("in generateSelfSignedTLSConf, certPrivateKey = %v\n", certPrivateKey)

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &certPrivateKey.PublicKey, certPrivateKey)
	if err != nil {
		return nil, err
	}
	log.Printf("in generateSelfSignedTLSConf, certBytes = %v\n", certBytes)

	certPEM := new(bytes.Buffer)
	err = pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		return nil, err
	}
	log.Printf("in generateSelfSignedTLSConf, certPEM = %v\n", certPEM)

	certPrivateKeyPEM := new(bytes.Buffer)
	err = pem.Encode(certPrivateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivateKey),
	})
	if err != nil {
		return nil, err
	}
	log.Printf("in generateSelfSignedTLSConf, certPrivateKeyPEM = %v\n", certPrivateKeyPEM)

	log.Printf("cert PEM = %s\n", certPEM.String())
	log.Printf("key PEM = %s\n", certPrivateKeyPEM.String())

	serverCert, err := tls.X509KeyPair(certPEM.Bytes(), certPrivateKeyPEM.Bytes())
	if err != nil {
		return nil, err
	}

	rootCAs := bestEffortSystemCertPool()
	rootCAs.AppendCertsFromPEM(certPEM.Bytes())

	m.k.StoreServerTLSConf(certPEM.String(), certPrivateKeyPEM.String(), AtlasNamespace)

	// TODO: guess it's a good idea to add certificates to the kubernetes secret
	//		and other servers will have to pull and use this certificate
	// TODO: main certificate should be added by workflowtrigger
	//		other servers should listen to secret and only start servers and clients
	//		when secret is reachable and cert is parsed and added to the pools
	// TODO: certificate distribution:
	//		workflowtrigger - generates and updates cert and key in kuber secret
	//		workfloworchestrator - should retrieve key and cert from secrets and listen for secret updates
	//		reposever - should retrieve key and cert from secrets and listen for secret updates
	//		command delegator - should retrieve key and cert from secrets and listen for secret updates
	//		client wrapper - should somehow fetch cert from workflow trigger
	//			could not retrieve cert and key from secrets as it is located in the different clusters
	//			also should somehow subscribe for cert and key updates
	//			as an option it could retrieve cert from workflow trigger api
	//			other option is to create an additional service which will manage tls for all other components
	//		cli - should receive cert to perform requests to the workflowtrigger server
	//			as an option it could send http request to the workflowtrigger and fetch cert
	//			after that cli client could be updated to perform https requests
	// TODO: do we need to secure communication with dbs? (kafka, redis)
	log.Printf("in generateSelfSignedTLSConf, serverCert = %v\n", serverCert)
	return &tls.Config{
		Certificates:             []tls.Certificate{serverCert},
		MinVersion:               tls.VersionTLS13,
		PreferServerCipherSuites: true,
		RootCAs:                  rootCAs,
	}, nil
}

func generateCertificateSerialNumber() (*big.Int, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	return rand.Int(rand.Reader, serialNumberLimit)
}
