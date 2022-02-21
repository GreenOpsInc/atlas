package apikeys

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/greenopsinc/util/kubernetesclient"
	v1 "k8s.io/api/core/v1"

	"github.com/greenopsinc/util/rand"
)

const ATLAS_NAMESPACE = "atlas"

type Client interface {
	Issue(name string) (string, error)
	GetOne(name string) (string, error)
	GetAll() map[string]string
	Verify(apikey string) (bool, error)
	Rotate(name string) (string, error)
	WatchApiKeys() error
	WatchApiKeySecret(name string, handler kubernetesclient.WatchSecretHandler) error
}

type client struct {
	apikeys map[string]string
	kclient kubernetesclient.KubernetesClient
}

func New(kclient kubernetesclient.KubernetesClient) Client {
	return &client{kclient: kclient}
}

func (c *client) WatchApiKeys() error {
	return c.kclient.WatchSecretData(context.Background(), kubernetesclient.ApikeysSecretName, ATLAS_NAMESPACE, func(_ kubernetesclient.SecretChangeType, secret *v1.Secret) {
		var keys map[string]string
		if err := json.Unmarshal(secret.Data[kubernetesclient.SecretsKeyName], &keys); err != nil {
			log.Println("failed to parse apikeys secret in kubernetes secrets watcher")
		}
		c.apikeys = keys
	})
}

func (c *client) WatchApiKeySecret(name string, handler kubernetesclient.WatchSecretHandler) error {
	return c.kclient.WatchSecretData(context.Background(), name, ATLAS_NAMESPACE, handler)
}

func (c *client) Issue(name string) (string, error) {
	apikey, err := rand.RandomString(64)
	if err != nil {
		return "", err
	}

	if !c.kclient.StoreApiKey(apikey, name, ATLAS_NAMESPACE) {
		return "", errors.New("failed to store apikey in secrets")
	}
	return apikey, nil
}

func (c *client) GetOne(name string) (string, error) {
	apikey := c.kclient.FetchApiKey(name, ATLAS_NAMESPACE)
	if apikey == "" {
		return "", errors.New("failed to fetch apikey from secrets")
	}
	return apikey, nil
}

func (c *client) GetAll() map[string]string {
	if c.apikeys != nil {
		return c.apikeys
	}
	keys := c.kclient.FetchApiKeys(ATLAS_NAMESPACE)
	c.apikeys = keys
	return keys
}

func (c *client) Verify(apikey string) (bool, error) {
	for _, val := range c.apikeys {
		if val == apikey {
			return true, nil
		}
	}
	return false, nil
}

func (c *client) Rotate(name string) (string, error) {
	return c.Issue(name)
}
