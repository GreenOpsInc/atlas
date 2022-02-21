package apikeysmanager

import (
	"fmt"

	"github.com/greenopsinc/util/apikeys"
	"github.com/greenopsinc/util/kubernetesclient"
)

const (
	workflowTriggerApiKeySecretName = "atlas-workflow-trigger-api-key"
	clientWrapperApiKeySecretName   = "atlas-client-wrapper-api-key"
	ApiKeyHeaderName                = "X-Api-Key"
)

type Manager interface {
	GenerateDefaultKeys() error
	GenerateClientWrapperApiKey(name string) (string, error)
	GetWorkflowTriggerApiKey() string
	RotateWorkflowTriggerApiKey() (string, error)
	RotateClientWrapperApiKey(name string) (string, error)
	VerifyRequest(apikey string) (bool, error)
	WatchApiKeys() error
}

type manager struct {
	client apikeys.Client
}

func New(kubernetesClient kubernetesclient.KubernetesClient) Manager {
	apikeysClient := apikeys.New(kubernetesClient)
	apikeysClient.GetAll()
	return &manager{client: apikeysClient}
}

func (m *manager) GenerateDefaultKeys() error {
	_, err := m.client.Issue(workflowTriggerApiKeySecretName)
	if err != nil {
		return err
	}
	_, err = m.client.Issue(clientWrapperApiKeySecretName)
	if err != nil {
		return err
	}
	return nil
}

func (m *manager) GenerateClientWrapperApiKey(clusterName string) (string, error) {
	secretName := fmt.Sprintf("atlas-%s-cluster-client-wrapper-api-key", clusterName)
	return m.client.Issue(secretName)
}

func (m *manager) GetWorkflowTriggerApiKey() string {
	keys := m.client.GetAll()
	return keys[workflowTriggerApiKeySecretName]
}

func (m *manager) RotateWorkflowTriggerApiKey() (string, error) {
	apikey, err := m.client.Rotate(workflowTriggerApiKeySecretName)
	if err != nil {
		return "", err
	}
	return apikey, nil
}

func (m *manager) RotateClientWrapperApiKey(clusterName string) (string, error) {
	return m.client.Rotate(getClientWrapperApiKeyName(clusterName))
}

func getClientWrapperApiKeyName(clusterName string) string {
	return fmt.Sprintf("atlas-%s-cluster-client-wrapper-api-key", clusterName)
}

func (m *manager) VerifyRequest(apikey string) (bool, error) {
	return m.client.Verify(apikey)
}

func (m *manager) WatchApiKeys() error {
	return m.client.WatchApiKeys()
}
