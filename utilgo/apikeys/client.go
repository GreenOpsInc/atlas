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

// create cluster method in wt will generate api key and send it to user
// single api key for cluster will secure cd and wt
// create refresh method which will rotate api key
// by default we will have default api key
// api key should only secure generate enpoint on wt
// api key should secure all endpoints on cd
// refresh api key method should be secured by argocd acess token
// refresh api key method should has cluster name\id parameters

// create api keys for each client: wt, cluster1 cw, cluster2 cw, etc.
// user should put cw in a secret and get `atlas-client-wrapper-apikey` `atlas-workflow-trigger-apikey`

// TODO: seems like we need to store two secrets: for workflow trigger and for command delegator

// TODO: how should we generate/rotate api keys (which api communication method to use)?
//		1. use http api call to service (not applicable as we need to somehow secure it)
//		2. use cli+workflowtrigger to generate/rotate secrets and api keys for services
//			guess thesis good option as cli and wt secured with argocd auth

// TODO: "When a user desires, they can query for a new token (on a per cluster basis), which should retire the cluster's old access token and provide a new one. This will require an API method."
//		In this case we cannot use a simple hash with secret and we have two options:
//			1. create a blacklist or whitelist of used/allowed api keys
//			2. rotate a secret each time when user wants to generate new api key

// TODO: where we should store secrets for api keys:
//		1. env vars
//		2. kubernetes secrets
//		3. redis

// TODO: use a single secret for all apikeys
//		add listener on wt and cd to listen for apikeys secret

const ATLAS_NAMESPACE = "atlas"

type Client interface {
	Issue(name string) (string, error)
	GetOne(name string) (string, error)
	GetAll() map[string]string
	Verify(apikey string) (bool, error)
	Rotate(name string) (string, error)
	WatchApiKeys() error
	WatchApiKeySecret(handler kubernetesclient.WatchSecretHandler) error
}

type client struct {
	apikeys map[string]string
	kclient kubernetesclient.KubernetesClient
}

func New(kclient kubernetesclient.KubernetesClient) Client {
	c := &client{kclient: kclient}
	keys := c.kclient.FetchApiKeys(ATLAS_NAMESPACE)
	if keys == nil {
		keys = make(map[string]string, 0)
	}
	c.apikeys = keys
	return c
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

func (c *client) WatchApiKeySecret(handler kubernetesclient.WatchSecretHandler) error {
	return c.kclient.WatchSecretData(context.Background(), kubernetesclient.ApikeysSecretName, ATLAS_NAMESPACE, handler)
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
