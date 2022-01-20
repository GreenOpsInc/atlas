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
		return x509.NewCertPool()
	}
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
	conf, err := m.getTLSClientConf(clientName)
	if err != nil {
		return nil, err
	}
	m.tlsClientConfigs[clientName] = conf
	return conf, nil
}

func (m *tlsManager) GetClientCertPEM(clientName ClientName) ([]byte, error) {
	if m.tlsClientCertPEM[clientName] != nil {
		return m.tlsClientCertPEM[clientName], nil
	}

	secretName, err := clientNameToSecretName(clientName)
	if err != nil {
		return nil, err
	}

	secret := m.k.FetchSecretData(string(secretName), NamespaceDefault)
	if secret == nil || len(secret.Data) == 0 {
		return nil, nil
	}
	m.tlsClientCertPEM[clientName] = secret.Data[TLSSecretCrtName]
	return secret.Data[TLSSecretCrtName], nil
}

func (m *tlsManager) GetKafkaTLSConf() (*tls.Config, error) {
	if m.tlsClientConfigs[ClientKafka] != nil {
		return m.tlsClientConfigs[ClientKafka], nil
	}

	secretName, err := clientNameToSecretName(ClientKafka)
	if err != nil {
		return nil, err
	}

	secret := m.k.FetchSecretData(string(secretName), NamespaceDefault)
	if secret == nil || len(secret.Data) == 0 {
		return &tls.Config{InsecureSkipVerify: true}, nil
	}
	rootCA := m.BestEffortSystemCertPool()
	rootCA.AppendCertsFromPEM(secret.Data[TLSSecretKafkaCrtName])
	m.tlsClientCertPEM[ClientKafka] = secret.Data[TLSSecretKafkaCrtName]
	return &tls.Config{RootCAs: rootCA}, nil
}

func (m *tlsManager) WatchServerTLSConf(serverName ClientName, handler func(conf *tls.Config, err error)) error {
	secretName, err := clientNameToSecretName(serverName)
	if err != nil {
		return err
	}

	ctx := context.Background()
	return m.k.WatchSecretData(ctx, string(secretName), NamespaceDefault, func(t kclient.SecretChangeType, secret *corev1.Secret) {
		var (
			config   *tls.Config
			err      error
			insecure bool
		)

		switch t {
		case kclient.SecretChangeTypeAdd:
			fallthrough
		case kclient.SecretChangeTypeUpdate:
			if m.tlsClientCertPEM[serverName] != nil && bytes.Compare(secret.Data[TLSSecretCrtName], m.tlsClientCertPEM[serverName]) == 0 {
				return
			}

			m.tlsClientCertPEM[serverName] = secret.Data[TLSSecretCrtName]
			config, err = m.generateTLSConfFromKeyPair(secret.Data[TLSSecretCrtName], secret.Data[TLSSecretKeyName])
			insecure = false
		case kclient.SecretChangeTypeDelete:
			if m.tlsConf == m.selfSignedConf {
				return
			}
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
	secretName, err := clientNameToSecretName(clientName)
	if err != nil {
		return err
	}

	ctx := context.Background()
	return m.k.WatchSecretData(ctx, string(secretName), NamespaceDefault, func(t kclient.SecretChangeType, secret *corev1.Secret) {
		switch t {
		case kclient.SecretChangeTypeAdd:
			fallthrough
		case kclient.SecretChangeTypeUpdate:
			if m.tlsClientCertPEM[clientName] != nil && bytes.Compare(secret.Data[TLSSecretCrtName], m.tlsClientCertPEM[clientName]) == 0 {
				return
			}

			m.tlsClientCertPEM[clientName] = secret.Data[TLSSecretCrtName]
			rootCA := m.BestEffortSystemCertPool()
			rootCA.AppendCertsFromPEM(secret.Data[TLSSecretCrtName])
			conf := &tls.Config{RootCAs: rootCA}
			m.tlsClientConfigs[clientName] = conf
			handler(conf, nil)
		case kclient.SecretChangeTypeDelete:
			if m.tlsClientCertPEM[clientName] == nil {
				return
			}

			delete(m.tlsClientCertPEM, clientName)
			delete(m.tlsClientConfigs, clientName)
			handler(&tls.Config{InsecureSkipVerify: true}, nil)
		}
	})
}

func (m *tlsManager) WatchClientTLSPEM(clientName ClientName, namespace string, handler func(certPEM []byte, err error)) error {
	secretName, err := clientNameToSecretName(clientName)
	if err != nil {
		return err
	}

	ctx := context.Background()
	return m.k.WatchSecretData(ctx, string(secretName), namespace, func(t kclient.SecretChangeType, secret *corev1.Secret) {
		switch t {
		case kclient.SecretChangeTypeAdd:
			fallthrough
		case kclient.SecretChangeTypeUpdate:
			crt := secret.Data[TLSSecretKafkaCrtName]
			if m.tlsClientCertPEM[clientName] != nil && bytes.Compare(crt, m.tlsClientCertPEM[clientName]) == 0 {
				return
			}

			m.tlsClientCertPEM[clientName] = secret.Data[TLSSecretCrtName]
			handler(secret.Data[TLSSecretCrtName], nil)
		case kclient.SecretChangeTypeDelete:
			if m.tlsClientCertPEM[clientName] == nil {
				return
			}

			delete(m.tlsClientCertPEM, clientName)
			handler(nil, nil)
		}
	})
}

func (m *tlsManager) WatchKafkaTLSConf(handler func(config *tls.Config, err error)) error {
	secretName, err := clientNameToSecretName(ClientKafka)
	if err != nil {
		return err
	}

	ctx := context.Background()
	return m.k.WatchSecretData(ctx, string(secretName), NamespaceDefault, func(t kclient.SecretChangeType, secret *corev1.Secret) {
		switch t {
		case kclient.SecretChangeTypeAdd:
			fallthrough
		case kclient.SecretChangeTypeUpdate:
			crt := secret.Data[TLSSecretKafkaCrtName]
			if m.tlsClientCertPEM[ClientKafka] != nil && bytes.Compare(crt, m.tlsClientCertPEM[ClientKafka]) == 0 {
				return
			}

			rootCA := m.BestEffortSystemCertPool()
			rootCA.AppendCertsFromPEM(crt)
			conf := &tls.Config{RootCAs: rootCA}
			m.tlsClientConfigs[ClientKafka] = conf
			m.tlsClientCertPEM[ClientKafka] = crt
			handler(conf, nil)
		case kclient.SecretChangeTypeDelete:
			if m.tlsClientCertPEM[ClientKafka] == nil {
				return
			}
			delete(m.tlsClientCertPEM, ClientKafka)
			delete(m.tlsClientConfigs, ClientKafka)
			m.tlsClientCertPEM[ClientKafka] = nil
			handler(nil, nil)
		}
	})
}

func (m *tlsManager) getTLSConf(serverName ClientName) (*tls.Config, error) {
	if m.tlsConf != nil {
		return m.tlsConf, nil
	}

	conf, err := m.getTLSConfFromSecrets(serverName)
	if err != nil {
		return nil, err
	}
	if conf != nil {
		conf.InsecureSkipVerify = false
		return conf, nil
	}

	conf, err = m.getSelfSignedTLSConf(serverName)
	if err != nil {
		return nil, err
	}

	conf.InsecureSkipVerify = true
	m.tlsConf = conf
	return conf, nil
}

func (m *tlsManager) getTLSClientConf(clientName ClientName) (*tls.Config, error) {
	if m.tlsClientConfigs[clientName] != nil {
		return m.tlsClientConfigs[clientName], nil
	}

	secretName, err := clientNameToSecretName(clientName)
	if err != nil {
		return nil, err
	}

	secret := m.k.FetchSecretData(string(secretName), NamespaceDefault)
	if secret == nil || len(secret.Data) == 0 {
		return &tls.Config{InsecureSkipVerify: true}, nil
	}

	rootCA := m.BestEffortSystemCertPool()
	rootCA.AppendCertsFromPEM(secret.Data[TLSSecretCrtName])
	m.tlsClientCertPEM[clientName] = secret.Data[TLSSecretCrtName]
	return &tls.Config{RootCAs: rootCA}, nil
}

func (m *tlsManager) getSelfSignedTLSConf(serverName ClientName) (*tls.Config, error) {
	if m.selfSignedConf != nil {
		return m.selfSignedConf, nil
	}

	conf, err := m.generateSelfSignedTLSConf(serverName)
	if err != nil {
		return nil, err
	}

	m.selfSignedConf = conf
	return conf, nil
}

func (m *tlsManager) getTLSConfFromSecrets(serverName ClientName) (*tls.Config, error) {
	secretName, err := clientNameToSecretName(serverName)
	if err != nil {
		return nil, err
	}
	secret := m.k.FetchSecretData(string(secretName), NamespaceDefault)
	if secret == nil || len(secret.Data) == 0 {
		return nil, nil
	}

	conf, err := m.generateTLSConfFromKeyPair(secret.Data[TLSSecretCrtName], secret.Data[TLSSecretKeyName])
	if err != nil {
		return nil, err
	}
	m.tlsClientCertPEM[serverName] = secret.Data[TLSSecretCrtName]
	return conf, nil
}

func (m *tlsManager) generateTLSConfFromKeyPair(certPEM []byte, keyPEM []byte) (*tls.Config, error) {
	c, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}
	rootCAs := m.BestEffortSystemCertPool()
	rootCAs.AppendCertsFromPEM(certPEM)
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

	m.tlsClientCertPEM[serverName] = certPEM.Bytes()
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
