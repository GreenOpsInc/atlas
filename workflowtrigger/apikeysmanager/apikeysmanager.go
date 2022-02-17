package apikeysmanager

import (
	"fmt"

	"github.com/greenopsinc/util/apikeys"
	"github.com/greenopsinc/util/kubernetesclient"
)

const (
	workflowTriggerApiKeySecretName = "atlas-workflow-trigger-api-key"
	clientWrapperApiKeySecretName   = "atlas-client-wrapper-api-key"
	ApiKeyHeaderName                = "api-key"
)

type Manager interface {
	GenerateDefaultKeys() error
	GenerateClientWrapperApiKey(name string) (string, error)
	GetWorkflowTriggerApiKey() (string, error)
	RotateWorkflowTriggerApiKey() (string, error)
	RotateClientWrapperApiKey(name string) (string, error)
	VerifyRequest(apikey string) bool
}

type manager struct {
	client                apikeys.Client
	workflowTriggerApiKey string
}

func New(kubernetesClient kubernetesclient.KubernetesClient) Manager {
	return &manager{client: apikeys.New(kubernetesClient)}
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

func (m *manager) GetWorkflowTriggerApiKey() (string, error) {
	apikey, err := m.client.Get(workflowTriggerApiKeySecretName)
	if err != nil {
		return "", err
	}
	if apikey != "" {
		return apikey, nil
	}
	apikey, err = m.client.Issue(workflowTriggerApiKeySecretName)
	if err != nil {
		return "", err
	}
	return apikey, nil
}

func (m *manager) RotateWorkflowTriggerApiKey() (string, error) {
	apikey, err := m.client.Rotate(workflowTriggerApiKeySecretName)
	if err != nil {
		return "", err
	}
	m.workflowTriggerApiKey = apikey
	return apikey, nil
}

func (m *manager) RotateClientWrapperApiKey(clusterName string) (string, error) {
	return m.client.Rotate(getClientWrapperApiKeyName(clusterName))
}

func getClientWrapperApiKeyName(clusterName string) string {
	return fmt.Sprintf("atlas-%s-cluster-client-wrapper-api-key", clusterName)
}

func (m *manager) VerifyRequest(apikey string) bool {
	return apikey == m.workflowTriggerApiKey
}
