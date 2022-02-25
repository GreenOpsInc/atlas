package kubernetesclient

import (
	"context"
	"github.com/greenopsinc/util/git"
	"github.com/greenopsinc/util/kubernetesclient"
	v1 "k8s.io/api/core/v1"
)

type SecretChangeType int8

type WatchSecretHandler func()

type MockKubernetesClient interface {
	StoreGitCred(gitCred git.GitCred, name string) bool
	FetchGitCred(name string) git.GitCred
}

func New() kubernetesclient.KubernetesClient {
	return MockKubernetesClientDriver{}
}

type MockKubernetesClientDriver struct {
}

func (m MockKubernetesClientDriver) FetchSecretData(name string, namespace string) *v1.Secret {
	//TODO implement me
	panic("implement me")
}

func (m MockKubernetesClientDriver) WatchSecretData(ctx context.Context, name string, namespace string, handler kubernetesclient.WatchSecretHandler) error {
	//TODO implement me
	panic("implement me")
}

func (m MockKubernetesClientDriver) StoreGitCred(gitCred git.GitCred, name string) bool {
	return true
}

func (m MockKubernetesClientDriver) FetchGitCred(name string) git.GitCred {
	//TODO implement me
	panic("implement me")
}
