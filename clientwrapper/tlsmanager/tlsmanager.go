package tlsmanager

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"log"

	kclient "greenops.io/client/kubernetesclient"
	corev1 "k8s.io/api/core/v1"
)

type Manager interface {
	BestEffortSystemCertPool() *x509.CertPool
	GetClientTLSConf(clientName ClientName) (*tls.Config, error)
	GetClientCertPEM(clientName ClientName) ([]byte, error)
	WatchClientTLSConf(clientName ClientName, handler func(conf *tls.Config, err error)) error
	WatchClientTLSPEM(clientName ClientName, handler func(certPEM []byte, err error)) error
}

type tlsManager struct {
	k                kclient.KubernetesClient
	tlsClientConfigs map[ClientName]*tls.Config
	tlsClientCertPEM map[ClientName][]byte
}

type TLSSecretName string

// TODO: currently those values are hardcoded, fetch them from config or somewhere else
const (
	NotValidSecretName            TLSSecretName = "not-valid"
	ClientWrapperTLSSecretName    TLSSecretName = "clientwrapper-tls"
	WorkflowTriggerTLSSecretName  TLSSecretName = "workflowtrigger-tls"
	CommandDelegatorTLSSecretName TLSSecretName = "commanddelegator-tls"
	ArgoCDRepoServerTLSSecretName TLSSecretName = "argocd-repo-server-tls"
	Namespace                     string        = "default"
)

type ClientName string

const (
	ClientWorkflowTrigger  ClientName = "workflowtrigger"
	ClientCommandDelegator ClientName = "commanddelegator"
	ClientArgoCDRepoServer ClientName = "argocdreposerver"
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
	if _, err := m.GetClientTLSConf(ClientWorkflowTrigger); err != nil {
		return err
	}
	if _, err := m.GetClientTLSConf(ClientCommandDelegator); err != nil {
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

func (m *tlsManager) GetClientTLSConf(clientName ClientName) (*tls.Config, error) {
	conf, err := m.getTLSClientConf(clientName)
	if err != nil {
		return nil, err
	}
	m.tlsClientConfigs[clientName] = conf
	return conf, nil
}

func (m *tlsManager) GetClientCertPEM(clientName ClientName) ([]byte, error) {
	log.Println("in GetClientCertPEM")
	if m.tlsClientCertPEM[clientName] != nil {
		return m.tlsClientCertPEM[clientName], nil
	}

	secretName, err := clientNameToSecretName(clientName)
	if err != nil {
		return nil, err
	}

	secret := m.k.FetchSecretData(string(secretName), Namespace)
	if secret == nil {
		return nil, nil
	}
	return secret[TLSSecretCrtName], nil
}

// TODO: check that this watcher is not trigger server reloading if cert is not changed
//		it could possibly happend on server start when cert is available and we also receiving secret change event
func (m *tlsManager) WatchClientTLSConf(clientName ClientName, handler func(conf *tls.Config, err error)) error {
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

func (m *tlsManager) WatchClientTLSPEM(clientName ClientName, handler func(certPEM []byte, err error)) error {
	log.Println("in WatchClientTLSPEM")
	secretName, err := clientNameToSecretName(clientName)
	if err != nil {
		return err
	}

	m.k.WatchSecretData(string(secretName), Namespace, func(t kclient.SecretChangeType, secret *corev1.Secret) {
		log.Printf("in WatchClientTLSPEM, event %v. data = %s\n", t, secret)
		switch t {
		case kclient.SecretChangeTypeAdd:
			fallthrough
		case kclient.SecretChangeTypeUpdate:
			handler(secret.Data[TLSSecretCrtName], nil)
		case kclient.SecretChangeTypeDelete:
			handler(nil, nil)
		}
	})
	return nil
}

func (m *tlsManager) getTLSClientConf(clientName ClientName) (*tls.Config, error) {
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

func clientNameToSecretName(clientName ClientName) (TLSSecretName, error) {
	switch clientName {
	case ClientWorkflowTrigger:
		return WorkflowTriggerTLSSecretName, nil
	case ClientCommandDelegator:
		return CommandDelegatorTLSSecretName, nil
	case ClientArgoCDRepoServer:
		return ArgoCDRepoServerTLSSecretName, nil
	default:
		return NotValidSecretName, errors.New("wrong client name provided to get client secret")
	}
}
