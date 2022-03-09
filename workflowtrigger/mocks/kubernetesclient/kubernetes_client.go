package kubernetesclient

import (
	"context"
	"fmt"
	"github.com/greenopsinc/util/kubernetesclient"
	"k8s.io/apimachinery/pkg/runtime"
	"log"
	"strings"
	"time"

	"github.com/greenopsinc/util/git"
	"github.com/greenopsinc/util/serializer"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	dynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/informers"
	normalclient "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

const (
	secretsKeyName   string = "data"
	gitCredNamespace string = "gitcred"
)

type SecretChangeType int8

const (
	SecretChangeTypeAdd    SecretChangeType = 1
	SecretChangeTypeUpdate SecretChangeType = 2
	SecretChangeTypeDelete SecretChangeType = 3
)

type TLSSecretData struct {
	Crt string `json:"crt"`
	Key string `json:"key"`
}

type MockKubernetesClient interface {
	StoreGitCred(gitCred git.GitCred, name string) bool
	FetchGitCred(name string) git.GitCred
	FetchSecretData(name string, namespace string) *v1.Secret
	WatchSecretData(ctx context.Context, name string, namespace string, handler kubernetesclient.WatchSecretHandler) error
}

type MockKubernetesClientDriver struct {
	client        *normalclient.Clientset
	dynamicClient *dynamic.FakeDynamicClient
}

//TODO: ALL functions should have a callee tag on them
func New() kubernetesclient.KubernetesClient {
	schema := runtime.NewScheme()
	normalClient := normalclient.NewSimpleClientset()
	dynamicClient := dynamic.NewSimpleDynamicClient(schema)
	var client kubernetesclient.KubernetesClient
	client = MockKubernetesClientDriver{normalClient, dynamicClient}
	return client
}

func (k MockKubernetesClientDriver) StoreGitCred(gitCred git.GitCred, name string) bool {
	var err error
	if gitCred == nil {
		err = k.storeSecret(nil, gitCredNamespace, name)
	} else {
		err = k.storeSecret(gitCred.(interface{}), gitCredNamespace, name)
	}
	if err != nil {
		return false
	}
	fmt.Println("here")
	return true
}

func (k MockKubernetesClientDriver) FetchGitCred(name string) git.GitCred {
	secret := k.readSecret(gitCredNamespace, name)
	if secret != nil {
		if val, ok := secret.StringData[secretsKeyName]; ok {
			return git.UnmarshallGitCredString(val)
		}
		return nil
	}
	return nil
}

func (k MockKubernetesClientDriver) FetchSecretData(name string, namespace string) *v1.Secret {
	secret := k.readSecret(namespace, name)
	if secret != nil {
		return secret
	}
	return nil
}

func (k MockKubernetesClientDriver) WatchSecretData(ctx context.Context, name string, namespace string, handler kubernetesclient.WatchSecretHandler) error {
	factory := informers.NewSharedInformerFactoryWithOptions(k.client, time.Second*30, informers.WithNamespace(namespace))
	informer := factory.Core().V1().Secrets().Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			handleSecretInformerEvent(obj, name, SecretChangeTypeAdd, handler)
		},
		UpdateFunc: func(_, newObj interface{}) {
			handleSecretInformerEvent(newObj, name, SecretChangeTypeUpdate, handler)
		},
		DeleteFunc: func(obj interface{}) {
			handleSecretInformerEvent(obj, name, SecretChangeTypeDelete, handler)
		},
	})
	err := informer.SetWatchErrorHandler(func(r *cache.Reflector, err error) {
		log.Println("failed to watch secret values: ", err)
	})
	if err != nil {
		return err
	}
	go informer.Run(ctx.Done())
	return nil
}

func handleSecretInformerEvent(obj interface{}, name string, t SecretChangeType, handler kubernetesclient.WatchSecretHandler) {
	secret, ok := obj.(*corev1.Secret)
	if !ok {
		return
	}
	if secret.Name != name {
		return
	}
	handler(kubernetesclient.SecretChangeType(t), secret)
}

func (k MockKubernetesClientDriver) storeSecret(object interface{}, namespace string, name string) error {
	var err error
	_, err = k.CheckAndCreateNamespace(namespace)
	if err != nil {
		return err
	}
	if object == nil {
		return k.deleteSecret(namespace, name)
	} else if k.readSecret(namespace, name) == nil {
		return k.createSecret(object, namespace, name)
	}
	return k.updateSecret(object, namespace, name)
}

func (k MockKubernetesClientDriver) createSecret(object interface{}, namespace string, name string) error {
	var err error
	secret := makeSecret(object, namespace, name)
	_, err = k.client.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		log.Printf("Error encountered while reading secret: %s", err)
	}
	return err
}

func (k MockKubernetesClientDriver) updateSecret(object interface{}, namespace string, name string) error {
	var err error
	secret := makeSecret(object, namespace, name)
	_, err = k.client.CoreV1().Secrets(namespace).Update(context.TODO(), secret, metav1.UpdateOptions{})
	if err != nil {
		log.Printf("Error encountered while reading secret: %s", err)
	}
	return err
}

func (k MockKubernetesClientDriver) readSecret(namespace string, name string) *corev1.Secret {
	secret, err := k.client.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if strings.HasSuffix(err.Error(), "not found") {
			return nil
		}
		log.Printf("Error encountered while reading secret: %s", err)
		return nil
	}
	return secret
}

func (k MockKubernetesClientDriver) deleteSecret(namespace string, name string) error {
	err := k.client.CoreV1().Secrets(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		log.Printf("Error encountered while deleting secret: %s", err)
	}
	return err
}

func (k MockKubernetesClientDriver) CheckAndCreateNamespace(namespace string) (string, error) {
	if namespace == "" {
		namespace = corev1.NamespaceDefault
	}
	if _, err := k.client.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{}); err == nil {
		return namespace, nil
	}
	newNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespace,
			Namespace: "",
		},
	}
	_, err := k.client.CoreV1().Namespaces().Create(context.TODO(), newNamespace, metav1.CreateOptions{})
	if err != nil && errors.IsAlreadyExists(err) {
		err = nil
	}
	if err != nil {
		log.Printf("Error encountered while creating namespace: %s", err)
	}
	return namespace, err
}

func makeSecret(object interface{}, namespace string, name string) *corev1.Secret {
	objectBytes := []byte(serializer.Serialize(object))
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		StringData: map[string]string{secretsKeyName: string(objectBytes)},
		Type:       "Opaque",
	}
}
