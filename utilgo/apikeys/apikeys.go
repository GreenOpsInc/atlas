package apikeys

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/greenopsinc/util/rand"
)

// create cluster method in wt will generate api key and send it to user
// single api key for cluster will secure cd and wt
// create refresh method which will rotate api key
// by default we will have default api key
// api key should only secure generate enpoint
// api key should secure all endpoints
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

func Generate(secret string) (string, error) {
	h := hmac.New(sha256.New, []byte(secret))
	data, err := rand.RandomString(64)
	if err != nil {
		return "", err
	}
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil)), nil
}

func Verify(hash, secret string) (bool, error) {
	sig, err := hex.DecodeString(hash)
	if err != nil {
		return false, err
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(secret))
	return hmac.Equal(sig, mac.Sum(nil)), nil
}
