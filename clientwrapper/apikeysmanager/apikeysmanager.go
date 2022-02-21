package apikeysmanager

import (
	"github.com/greenopsinc/util/apikeys"
	"github.com/greenopsinc/util/kubernetesclient"
	v1 "k8s.io/api/core/v1"
)

type Manager interface {
	GetApiKey() string
	WatchApiKey() error
}

type manager struct {
	apikey     string
	secretName string
	client     apikeys.Client
}

func New(secretName string, kubernetesClient kubernetesclient.KubernetesClient) (Manager, error) {
	m := &manager{secretName: secretName, client: apikeys.New(kubernetesClient)}
	apikey, err := m.client.GetOne(m.secretName)
	if err != nil {
		return nil, err
	}
	m.apikey = apikey
	return m, nil
}

func (m *manager) GetApiKey() string {
	return m.apikey
}

func (m *manager) WatchApiKey() error {
	return m.client.WatchApiKeySecret(m.secretName, func(_ kubernetesclient.SecretChangeType, secret *v1.Secret) {
		m.apikey = string(secret.Data[kubernetesclient.SecretsKeyName])
	})
}
