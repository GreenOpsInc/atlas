package tlsmanager

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"log"
	"math/big"
	"net"
	"time"

	"greenops.io/workflowtrigger/client"
	kclient "greenops.io/workflowtrigger/kubernetesclient"
	corev1 "k8s.io/api/core/v1"
)

type Manager interface {
	BestEffortSystemCertPool() *x509.CertPool
	GetServerTLSConf() (*tls.Config, error)
	GetClientTLSConf(clientName client.ClientName) (*tls.Config, error)
	WatchServerTLSConf(handler func(conf *tls.Config, err error))
	WatchClientTLSConf(clientName client.ClientName, handler func(conf *tls.Config, err error)) error
}

type tlsManager struct {
	k                kclient.KubernetesClient
	tlsConf          *tls.Config
	selfSignedConf   *tls.Config
	tlsClientConfigs map[client.ClientName]*tls.Config
}

type TLSSecretName string

// TODO: currently those values are hardcoded, fetch them from config or somewhere else
const (
	NotValidSecretName            TLSSecretName = "not-valid"
	WorkflowTriggerTLSSecretName  TLSSecretName = "workflowtrigger-tls"
	RepoServerTLSSecretName       TLSSecretName = "pipelinereposerver-tls"
	CommandDelegatorTLSSecretName TLSSecretName = "commanddelegator-tls"
	ArgoCDRepoServerTLSSecretName TLSSecretName = "argocd-repo-server-tls"
	Namespace                     string        = "default"
)

const (
	TLSSecretCrtName = "tls.crt"
	TLSSecretKeyName = "tls.key"
)

func New(k kclient.KubernetesClient) (Manager, error) {
	m := &tlsManager{k: k}
	if err := m.initConfigs(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *tlsManager) initConfigs() error {
	if _, err := m.GetServerTLSConf(); err != nil {
		return err
	}
	if _, err := m.GetClientTLSConf(client.ClientRepoServer); err != nil {
		return err
	}
	if _, err := m.GetClientTLSConf(client.ClientCommandDelegator); err != nil {
		return err
	}
	if _, err := m.GetClientTLSConf(client.ClientArgoCDRepoServer); err != nil {
		return err
	}
	return nil
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

func (m *tlsManager) GetServerTLSConf() (*tls.Config, error) {
	conf, err := m.getTLSConf()
	if err != nil {
		return nil, err
	}
	m.tlsConf = conf
	return conf, nil
}

func (m *tlsManager) GetClientTLSConf(clientName client.ClientName) (*tls.Config, error) {
	conf, err := m.getTLSClientConf(clientName)
	if err != nil {
		return nil, err
	}
	m.tlsClientConfigs[clientName] = conf
	return conf, nil
}

// TODO: check that this watcher is not trigger server reloading if cert is not changed
//		it could possibly happend on server start when cert is available and we also receiving secret change event
func (m *tlsManager) WatchServerTLSConf(handler func(conf *tls.Config, err error)) {
	log.Println("in WatchServerTLSConf")
	m.k.WatchSecretData(string(WorkflowTriggerTLSSecretName), Namespace, func(t kclient.SecretChangeType, secret *corev1.Secret) {
		log.Printf("in WatchServerTLSConf, event %v. data = %s\n", t, secret)
		var (
			config   *tls.Config
			err      error
			insecure bool
		)

		switch t {
		case kclient.SecretChangeTypeAdd:
			fallthrough
		case kclient.SecretChangeTypeUpdate:
			log.Printf("in WatchServerTLSConf, secret data = %v\n", secret.Data)
			config, err = m.generateTLSConfFromKeyPair(secret.Data[TLSSecretCrtName], secret.Data[TLSSecretCrtName])
			log.Printf("in WatchServerTLSConf, tlsConf = %v\n", config)
			insecure = false
		case kclient.SecretChangeTypeDelete:
			config, err = m.getSelfSignedTLSConf()
			insecure = true
		}

		if err != nil {
			handler(nil, err)
			return
		}

		config.InsecureSkipVerify = insecure
		m.tlsConf = config
		handler(config, nil)
	})
}

// TODO: check that this watcher is not trigger server reloading if cert is not changed
//		it could possibly happend on server start when cert is available and we also receiving secret change event
func (m *tlsManager) WatchClientTLSConf(clientName client.ClientName, handler func(conf *tls.Config, err error)) error {
	log.Println("in WatchClientTLSConf")
	secretName, err := clientNameToSecretName(clientName)
	if err != nil {
		return err
	}

	m.k.WatchSecretData(string(secretName), Namespace, func(t kclient.SecretChangeType, secret *corev1.Secret) {
		log.Printf("in WatchClientTLSConf, event %v. data = %s\n", t, secret)
		switch t {
		case kclient.SecretChangeTypeAdd:
			fallthrough
		case kclient.SecretChangeTypeUpdate:
			rootCA := m.BestEffortSystemCertPool()
			rootCA.AppendCertsFromPEM(secret.Data[TLSSecretCrtName])
			handler(&tls.Config{RootCAs: rootCA}, nil)
		case kclient.SecretChangeTypeDelete:
			handler(&tls.Config{InsecureSkipVerify: true}, nil)
		}
	})
	return nil
}

// TODO: need to create a general method for get to get delegator, repo server and wrapper secrets
func (m *tlsManager) getTLSConf() (*tls.Config, error) {
	log.Println("in GetServerTLSConf")
	if m.tlsConf != nil {
		return m.tlsConf, nil
	}

	conf, err := m.getTLSConfFromSecrets()
	log.Printf("in GetServerTLSConf, tlsConf = %v\n", conf)
	if err != nil {
		return nil, err
	}
	if conf != nil {
		log.Println("CERT FOUND IN SECRETS")
		conf.InsecureSkipVerify = false
		return conf, nil
	}

	log.Println("in GetServerTLSConf, before getSelfSignedTLSConf")
	conf, err = m.getSelfSignedTLSConf()
	if err != nil {
		return nil, err
	}

	conf.InsecureSkipVerify = true
	return conf, nil
}

func (m *tlsManager) getTLSClientConf(clientName client.ClientName) (*tls.Config, error) {
	log.Println("in GetServerTLSConf")
	if m.tlsClientConfigs[clientName] != nil {
		return m.tlsClientConfigs[clientName], nil
	}

	secretName, err := clientNameToSecretName(clientName)
	if err != nil {
		return nil, err
	}

	secret := m.k.FetchSecretData(string(secretName), Namespace)
	if secret == nil {
		return &tls.Config{InsecureSkipVerify: true}, nil
	}

	rootCA := m.BestEffortSystemCertPool()
	rootCA.AppendCertsFromPEM(secret[TLSSecretCrtName])
	return &tls.Config{RootCAs: rootCA}, nil
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
	log.Printf("in getSelfSignedTLSConf, tlsConf = %v\n", conf)

	m.selfSignedConf = conf
	return conf, nil
}

func (m *tlsManager) getTLSConfFromSecrets() (*tls.Config, error) {
	log.Println("in getTLSConfFromSecrets")
	secret := m.k.FetchSecretData(string(WorkflowTriggerTLSSecretName), Namespace)
	log.Println("in getTLSConfFromSecrets, secret: ", secret)
	if secret == nil {
		return nil, nil
	}

	conf, err := m.generateTLSConfFromKeyPair(secret[TLSSecretCrtName], secret[TLSSecretKeyName])
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func (m *tlsManager) generateTLSConfFromKeyPair(cert []byte, key []byte) (*tls.Config, error) {
	log.Printf("in generateTLSConfFromKeyPair cert = %s, key = %s\n", string(cert), string(key))
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

	rootCAs := m.BestEffortSystemCertPool()
	rootCAs.AppendCertsFromPEM(certPEM.Bytes())

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

func clientNameToSecretName(clientName client.ClientName) (TLSSecretName, error) {
	switch clientName {
	case client.ClientRepoServer:
		return RepoServerTLSSecretName, nil
	case client.ClientCommandDelegator:
		return CommandDelegatorTLSSecretName, nil
	case client.ClientArgoCDRepoServer:
		return ArgoCDRepoServerTLSSecretName, nil
	default:
		return NotValidSecretName, errors.New("wrong client name provided to get client secret")
	}
}
