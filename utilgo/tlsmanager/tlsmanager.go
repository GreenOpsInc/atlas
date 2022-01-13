package tlsmanager

import (
	"bytes"
	"context"
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

	kclient "github.com/greenopsinc/util/kubernetesclient"
	corev1 "k8s.io/api/core/v1"
)

type Manager interface {
	BestEffortSystemCertPool() *x509.CertPool
	GetServerTLSConf(serverName ClientName) (*tls.Config, error)
	GetClientTLSConf(clientName ClientName) (*tls.Config, error)
	GetClientCertPEM(clientName ClientName) ([]byte, error)
	GetKafkaTLSConf() (*tls.Config, error)
	WatchServerTLSConf(serverName ClientName, handler func(conf *tls.Config, err error)) error
	WatchClientTLSConf(clientName ClientName, handler func(conf *tls.Config, err error)) error
	WatchClientTLSPEM(clientName ClientName, namespace string, handler func(certPEM []byte, err error)) error
	WatchKafkaTLSConf(handler func(conf *tls.Config, err error)) error
}

type tlsManager struct {
	k                kclient.KubernetesClient
	tlsConf          *tls.Config
	selfSignedConf   *tls.Config
	tlsClientConfigs map[ClientName]*tls.Config
	tlsClientCertPEM map[ClientName][]byte
}

const (
	NamespaceDefault string = "default"
	NamespaceArgoCD  string = "argocd"
	CompanyName      string = "Atlas"
	CountryName      string = "US"
)

type TLSSecretName string

const (
	NotValidSecretName            TLSSecretName = "not-valid"
	WorkflowTriggerTLSSecretName  TLSSecretName = "workflowtrigger-tls"
	ClientWrapperTLSSecretName    TLSSecretName = "clientwrapper-tls"
	RepoServerTLSSecretName       TLSSecretName = "pipelinereposerver-tls"
	CommandDelegatorTLSSecretName TLSSecretName = "commanddelegator-tls"
	ArgoCDRepoServerTLSSecretName TLSSecretName = "argocd-repo-server-tls"
	KafkaTLSSecretName            TLSSecretName = "kafka-tls"
)

type ClientName string

const (
	ClientRepoServer       ClientName = "reposerver"
	ClientWorkflowTrigger  ClientName = "workflowtrigger"
	ClientClientWrapper    ClientName = "clientwrapper"
	ClientCommandDelegator ClientName = "commanddelegator"
	ClientArgoCDRepoServer ClientName = "argocdreposerver"
	ClientKafka            ClientName = "kafka"
)

type DNSName string

const (
	NotValidDNSName     DNSName = "not-valid"
	DNSLocalhost        DNSName = "localhost"
	DNSRepoServer       DNSName = "reposerver.default.svc.cluster.local"
	DNSWorkflowTrigger  DNSName = "workflowtrigger.default.svc.cluster.local"
	DNSCommandDelegator DNSName = "commanddelegator.default.svc.cluster.local"
	DNSArgoCDRepoServer DNSName = "argocd-server.argocd.svc.cluster.local"
)

const (
	TLSSecretCrtName      = "tls.crt"
	TLSSecretKeyName      = "tls.key"
	TLSSecretKafkaCrtName = "kafka.cert.pem"
)

func New(k kclient.KubernetesClient) Manager {
	m := &tlsManager{k: k}
	m.tlsClientConfigs = make(map[ClientName]*tls.Config)
	m.tlsClientCertPEM = make(map[ClientName][]byte)
	return m
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

func (m *tlsManager) GetServerTLSConf(serverName ClientName) (*tls.Config, error) {
	conf, err := m.getTLSConf(serverName)
	if err != nil {
		return nil, err
	}
	m.tlsConf = conf
	return conf, nil
}

func (m *tlsManager) GetClientTLSConf(clientName ClientName) (*tls.Config, error) {
	log.Println("in GetClientTLSConf ", clientName)
	conf, err := m.getTLSClientConf(clientName)
	log.Printf("received client conf of client %s. err = %s", clientName, err)
	if err != nil {
		return nil, err
	}
	log.Printf("assigninng client conf to map, map = %v", m.tlsClientConfigs)
	m.tlsClientConfigs[clientName] = conf
	return conf, nil
}

func (m *tlsManager) GetClientCertPEM(clientName ClientName) ([]byte, error) {
	log.Println("in GetClientCertPEM")
	if m.tlsClientCertPEM[clientName] != nil {
		log.Printf("in GetClientCertPEM %s client cer pem found in m.tlsClientCertPEM[clientName] map: %s", clientName, m.tlsClientCertPEM[clientName])
		return m.tlsClientCertPEM[clientName], nil
	}

	secretName, err := clientNameToSecretName(clientName)
	if err != nil {
		return nil, err
	}

	secret := m.k.FetchSecretData(string(secretName), NamespaceDefault)
	if secret == nil || len(secret.Data) == 0 {
		log.Println("in GetClientCertPEM secret is nil or data len == 0")
		return nil, nil
	}
	return secret.Data[TLSSecretCrtName], nil
}

func (m *tlsManager) GetKafkaTLSConf() (*tls.Config, error) {
	log.Println("in GetKafkaCertPEM")
	if m.tlsClientConfigs[ClientKafka] != nil {
		log.Printf("in GetKafkaCertPEM %s client cer pem found in m.tlsClientConfigs[clientName] map: %s", ClientKafka, m.tlsClientConfigs[ClientKafka])
		return m.tlsClientConfigs[ClientKafka], nil
	}

	secretName, err := clientNameToSecretName(ClientKafka)
	if err != nil {
		return nil, err
	}

	secret := m.k.FetchSecretData(string(secretName), NamespaceDefault)
	if secret == nil || len(secret.Data) == 0 {
		log.Println("in GetClientCertPEM secret is nil or data len == 0")
		return nil, nil
	}
	rootCA := m.BestEffortSystemCertPool()
	rootCA.AppendCertsFromPEM(secret.Data[TLSSecretKafkaCrtName])
	log.Printf("in GetKafkaTLSConf appending %v to root ca", secret.Data[TLSSecretKafkaCrtName])
	return &tls.Config{RootCAs: rootCA}, nil
}

func (m *tlsManager) WatchServerTLSConf(serverName ClientName, handler func(conf *tls.Config, err error)) error {
	log.Println("in WatchServerTLSConf")
	secretName, err := clientNameToSecretName(serverName)
	if err != nil {
		return err
	}

	ctx := context.Background()
	return m.k.WatchSecretData(ctx, string(secretName), NamespaceDefault, func(t kclient.SecretChangeType, secret *corev1.Secret) {
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
			config, err = m.generateTLSConfFromKeyPair(secret.Data[TLSSecretCrtName], secret.Data[TLSSecretKeyName])
			log.Printf("in WatchServerTLSConf, tlsConf = %v\n", config)
			insecure = false
		case kclient.SecretChangeTypeDelete:
			config, err = m.getSelfSignedTLSConf(serverName)
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

func (m *tlsManager) WatchClientTLSConf(clientName ClientName, handler func(conf *tls.Config, err error)) error {
	log.Println("in WatchClientTLSConf")
	secretName, err := clientNameToSecretName(clientName)
	if err != nil {
		return err
	}

	ctx := context.Background()
	return m.k.WatchSecretData(ctx, string(secretName), NamespaceDefault, func(t kclient.SecretChangeType, secret *corev1.Secret) {
		log.Printf("in WatchClientTLSConf, event %v. data = %s\n", t, secret)
		switch t {
		case kclient.SecretChangeTypeAdd:
			fallthrough
		case kclient.SecretChangeTypeUpdate:
			rootCA := m.BestEffortSystemCertPool()
			rootCA.AppendCertsFromPEM(secret.Data[TLSSecretCrtName])
			conf := &tls.Config{RootCAs: rootCA}
			m.tlsClientConfigs[clientName] = conf
			handler(conf, nil)
		case kclient.SecretChangeTypeDelete:
			delete(m.tlsClientConfigs, clientName)
			handler(&tls.Config{InsecureSkipVerify: true}, nil)
		}
	})
}

func (m *tlsManager) WatchClientTLSPEM(clientName ClientName, namespace string, handler func(certPEM []byte, err error)) error {
	log.Println("in WatchClientTLSPEM")
	secretName, err := clientNameToSecretName(clientName)
	if err != nil {
		return err
	}

	ctx := context.Background()
	return m.k.WatchSecretData(ctx, string(secretName), namespace, func(t kclient.SecretChangeType, secret *corev1.Secret) {
		log.Printf("in WatchClientTLSPEM, event %v. data = %s\n", t, secret)
		switch t {
		case kclient.SecretChangeTypeAdd:
			fallthrough
		case kclient.SecretChangeTypeUpdate:
			m.tlsClientCertPEM[clientName] = secret.Data[TLSSecretCrtName]
			handler(secret.Data[TLSSecretCrtName], nil)
		case kclient.SecretChangeTypeDelete:
			delete(m.tlsClientCertPEM, clientName)
			handler(nil, nil)
		}
	})
}

func (m *tlsManager) WatchKafkaTLSConf(handler func(config *tls.Config, err error)) error {
	log.Println("in WatchClientTLSConf")
	secretName, err := clientNameToSecretName(ClientKafka)
	if err != nil {
		return err
	}

	ctx := context.Background()
	return m.k.WatchSecretData(ctx, string(secretName), NamespaceDefault, func(t kclient.SecretChangeType, secret *corev1.Secret) {
		log.Printf("in WatchClientTLSConf, event %v. data = %s\n", t, secret)
		switch t {
		case kclient.SecretChangeTypeAdd:
			fallthrough
		case kclient.SecretChangeTypeUpdate:
			rootCA := m.BestEffortSystemCertPool()
			rootCA.AppendCertsFromPEM(secret.Data[TLSSecretKafkaCrtName])
			conf := &tls.Config{RootCAs: rootCA}
			m.tlsClientConfigs[ClientKafka] = conf
			handler(conf, nil)
		case kclient.SecretChangeTypeDelete:
			delete(m.tlsClientConfigs, ClientKafka)
			handler(nil, nil)
		}
	})
}

func (m *tlsManager) getTLSConf(serverName ClientName) (*tls.Config, error) {
	log.Println("in GetServerTLSConf")
	if m.tlsConf != nil {
		return m.tlsConf, nil
	}

	conf, err := m.getTLSConfFromSecrets(serverName)
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
	conf, err = m.getSelfSignedTLSConf(serverName)
	if err != nil {
		return nil, err
	}

	conf.InsecureSkipVerify = true
	return conf, nil
}

func (m *tlsManager) getTLSClientConf(clientName ClientName) (*tls.Config, error) {
	log.Println("in getTLSClientConf")
	if m.tlsClientConfigs[clientName] != nil {
		return m.tlsClientConfigs[clientName], nil
	}

	secretName, err := clientNameToSecretName(clientName)
	log.Printf("in getTLSClientConf secretName = %s", secretName)
	if err != nil {
		return nil, err
	}

	secret := m.k.FetchSecretData(string(secretName), NamespaceDefault)
	log.Printf("in getTLSClientConf secret = %s", secret)
	if secret == nil || len(secret.Data) == 0 {
		log.Printf("in getTLSClientConf for %s secret is nil, setting insecure to true", string(clientName))
		return &tls.Config{InsecureSkipVerify: true}, nil
	}
	log.Printf("in getTLSClientConf for %s secret is not nil, setting insecure to false", string(clientName))

	log.Println("in getTLSClientConf before getting root ca")
	rootCA := m.BestEffortSystemCertPool()
	rootCA.AppendCertsFromPEM(secret.Data[TLSSecretCrtName])
	log.Printf("in getTLSClientConf appending %v to root ca", secret.Data[TLSSecretCrtName])
	return &tls.Config{RootCAs: rootCA}, nil
}

func (m *tlsManager) getSelfSignedTLSConf(serverName ClientName) (*tls.Config, error) {
	log.Println("in getSelfSignedTLSConf")
	if m.selfSignedConf != nil {
		return m.selfSignedConf, nil
	}

	conf, err := m.generateSelfSignedTLSConf(serverName)
	if err != nil {
		return nil, err
	}
	log.Printf("in getSelfSignedTLSConf, tlsConf = %v\n", conf)

	m.selfSignedConf = conf
	return conf, nil
}

func (m *tlsManager) getTLSConfFromSecrets(serverName ClientName) (*tls.Config, error) {
	log.Println("in getTLSConfFromSecrets")
	secretName, err := clientNameToSecretName(serverName)
	log.Printf("in getTLSConfFromSecrets secretName = %s", secretName)
	if err != nil {
		return nil, err
	}
	secret := m.k.FetchSecretData(string(secretName), NamespaceDefault)
	log.Printf("in getTLSConfFromSecrets secret = %s", secret)
	log.Println("in getTLSConfFromSecrets, secret: ", secret)
	if secret == nil || len(secret.Data) == 0 {
		return nil, nil
	}

	log.Printf("in getTLSConfFromSecrets received secret, data = %v", secret.Data)
	log.Printf("in getTLSConfFromSecrets received secret, cert = %s, key = %s", secret.Data[TLSSecretCrtName], secret.Data[TLSSecretKeyName])

	conf, err := m.generateTLSConfFromKeyPair(secret.Data[TLSSecretCrtName], secret.Data[TLSSecretKeyName])
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func (m *tlsManager) generateTLSConfFromKeyPair(certPEM []byte, keyPEM []byte) (*tls.Config, error) {
	log.Printf("in generateTLSConfFromKeyPair cert = %s, key = %s\n", string(certPEM), string(keyPEM))
	c, err := tls.X509KeyPair(certPEM, keyPEM)
	log.Printf("in generateTLSConfFromKeyPair c = %v\n", c)
	if err != nil {
		log.Println("error in generateTLSConfFromKeyPair: ", err.Error())
		return nil, err
	}
	log.Println("in generateTLSConfFromKeyPair getting pool")
	rootCAs := m.BestEffortSystemCertPool()
	log.Println("in generateTLSConfFromKeyPair appending cert to pool")
	rootCAs.AppendCertsFromPEM(certPEM)
	log.Println("in generateTLSConfFromKeyPair returning tls conf")
	return &tls.Config{
		Certificates:             []tls.Certificate{c},
		MinVersion:               tls.VersionTLS13,
		PreferServerCipherSuites: true,
		RootCAs:                  rootCAs,
	}, nil
}

func (m *tlsManager) generateSelfSignedTLSConf(serverName ClientName) (*tls.Config, error) {
	rootCAs := m.BestEffortSystemCertPool()
	certSerialNumber, err := generateCertificateSerialNumber()
	if err != nil {
		return nil, err
	}
	dnsName, err := clientNameToDNSName(serverName)
	if err != nil {
		return nil, err
	}
	ca, caPrivKey, err := createCA(rootCAs)
	if err != nil {
		return nil, err
	}

	cert := &x509.Certificate{
		SerialNumber: certSerialNumber,
		Subject: pkix.Name{
			Organization: []string{CompanyName},
			Country:      []string{CountryName},
		},
		DNSNames:     []string{string(dnsName), string(DNSLocalhost)},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
	}

	certPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivateKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, err
	}

	certPEM := new(bytes.Buffer)
	err = pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		return nil, err
	}

	certPrivateKeyPEM := new(bytes.Buffer)
	err = pem.Encode(certPrivateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivateKey),
	})
	if err != nil {
		return nil, err
	}

	serverCert, err := tls.X509KeyPair(certPEM.Bytes(), certPrivateKeyPEM.Bytes())
	if err != nil {
		return nil, err
	}

	log.Println("cert = ", certPEM)
	log.Println("key = ", certPrivateKeyPEM)

	log.Printf("in generateSelfSignedTLSConf, serverCert = %v\n", serverCert)
	return &tls.Config{
		Certificates:             []tls.Certificate{serverCert},
		MinVersion:               tls.VersionTLS13,
		InsecureSkipVerify:       true,
		PreferServerCipherSuites: true,
	}, nil
}

func createCA(rootCAs *x509.CertPool) (*x509.Certificate, *rsa.PrivateKey, error) {
	caSerialNumber, err := generateCertificateSerialNumber()
	if err != nil {
		return nil, nil, err
	}
	ca := &x509.Certificate{
		SerialNumber: caSerialNumber,
		Subject: pkix.Name{
			Organization: []string{CompanyName},
			Country:      []string{CountryName},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	caPEM := new(bytes.Buffer)
	if err = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	}); err != nil {
		return nil, nil, err
	}
	caPrivKeyPEM := new(bytes.Buffer)
	if err = pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	}); err != nil {
		return nil, nil, err
	}

	rootCAs.AppendCertsFromPEM(caPEM.Bytes())
	return ca, caPrivKey, nil
}

func generateCertificateSerialNumber() (*big.Int, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	return rand.Int(rand.Reader, serialNumberLimit)
}

func clientNameToSecretName(clientName ClientName) (TLSSecretName, error) {
	switch clientName {
	case ClientRepoServer:
		return RepoServerTLSSecretName, nil
	case ClientWorkflowTrigger:
		return WorkflowTriggerTLSSecretName, nil
	case ClientClientWrapper:
		return ClientWrapperTLSSecretName, nil
	case ClientCommandDelegator:
		return CommandDelegatorTLSSecretName, nil
	case ClientArgoCDRepoServer:
		return ArgoCDRepoServerTLSSecretName, nil
	case ClientKafka:
		return KafkaTLSSecretName, nil
	default:
		return NotValidSecretName, errors.New("wrong client name provided to get client secret")
	}
}

func clientNameToDNSName(clientName ClientName) (DNSName, error) {
	switch clientName {
	case ClientRepoServer:
		return DNSRepoServer, nil
	case ClientWorkflowTrigger:
		return DNSWorkflowTrigger, nil
	case ClientCommandDelegator:
		return DNSCommandDelegator, nil
	case ClientArgoCDRepoServer:
		return DNSArgoCDRepoServer, nil
	default:
		return NotValidDNSName, errors.New("wrong client name provided to get client DNS name")
	}
}
