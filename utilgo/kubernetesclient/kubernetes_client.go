package kubernetesclient

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/greenopsinc/util/git"
	"github.com/greenopsinc/util/serializer"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const (
	gitCredNamespace  = "gitcred"
	SecretsKeyName    = "data"
	ApikeysSecretName = "atlas-apikeys"
)

type SecretChangeType int8

const (
	SecretChangeTypeAdd    SecretChangeType = 1
	SecretChangeTypeUpdate SecretChangeType = 2
	SecretChangeTypeDelete SecretChangeType = 3
)

type WatchSecretHandler func(action SecretChangeType, secret *corev1.Secret)

type TLSSecretData struct {
	Crt string `json:"crt"`
	Key string `json:"key"`
}

type KubernetesClient interface {
	StoreGitCred(gitCred git.GitCred, name string) bool
	StoreApiKey(apikey string, name string, namespace string) bool
	FetchGitCred(name string) git.GitCred
	FetchApiKey(name string, namespace string) string
	FetchApiKeys(namespace string) map[string]string
	FetchSecretData(name string, namespace string) *v1.Secret
	WatchSecretData(ctx context.Context, name string, namespace string, handler WatchSecretHandler) error
}

type KubernetesClientDriver struct {
	client        *kubernetes.Clientset
	dynamicClient dynamic.Interface
}

//TODO: ALL functions should have a callee tag on them
func New() KubernetesClient {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	normalClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	var client KubernetesClient
	client = KubernetesClientDriver{normalClient, dynamicClient}
	return client
}

func (k KubernetesClientDriver) StoreGitCred(gitCred git.GitCred, name string) bool {
	var err error
	if gitCred == nil {
		err = k.storeSecret(nil, gitCredNamespace, name)
	} else {
		err = k.storeSecret(gitCred.(interface{}), gitCredNamespace, name)
	}
	if err != nil {
		return false
	}
	return true
}

func (k KubernetesClientDriver) StoreApiKey(apikey string, name string, namespace string) bool {
	var apikeys map[string]string
	apikeysSecret := k.readSecret(namespace, ApikeysSecretName)
	if apikeysSecret == nil || apikeysSecret.Data == nil || apikeysSecret.Data[SecretsKeyName] == nil {
		apikeys = make(map[string]string, 1)
	} else {
		if err := json.Unmarshal(apikeysSecret.Data[SecretsKeyName], &apikeys); err != nil {
			return false
		}
	}

	apikeys[name] = apikey
	err := k.storeSecret(apikeys, namespace, name)
	return err == nil
}

func (k KubernetesClientDriver) FetchGitCred(name string) git.GitCred {
	secret := k.readSecret(gitCredNamespace, name)
	if secret != nil {
		if val, ok := secret.StringData[SecretsKeyName]; ok {
			return git.UnmarshallGitCredString(val)
		}
		return nil
	}
	return nil
}

func (k KubernetesClientDriver) FetchApiKeys(namespace string) map[string]string {
	secret := k.readSecret(namespace, ApikeysSecretName)
	if secret == nil {
		return nil
	}
	val, ok := secret.StringData[SecretsKeyName]
	if !ok {
		return nil
	}
	var res map[string]string
	if err := json.Unmarshal([]byte(val), &res); err != nil {
		return nil
	}
	return res
}

func (k KubernetesClientDriver) FetchApiKey(name string, namespace string) string {
	secret := k.readSecret(namespace, name)
	if secret == nil {
		return ""
	}
	if val, ok := secret.StringData[SecretsKeyName]; ok {
		return val
	}
	return ""
}

func (k KubernetesClientDriver) FetchSecretData(name string, namespace string) *v1.Secret {
	secret := k.readSecret(namespace, name)
	if secret != nil {
		return secret
	}
	return nil
}

func (k KubernetesClientDriver) WatchSecretData(ctx context.Context, name string, namespace string, handler WatchSecretHandler) error {
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

func handleSecretInformerEvent(obj interface{}, name string, t SecretChangeType, handler WatchSecretHandler) {
	secret, ok := obj.(*corev1.Secret)
	if !ok {
		return
	}
	if secret.Name != name {
		return
	}
	handler(t, secret)
}

func (k KubernetesClientDriver) storeSecret(object interface{}, namespace string, name string) error {
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

func (k KubernetesClientDriver) createSecret(object interface{}, namespace string, name string) error {
	var err error
	secret := makeSecret(object, namespace, name)
	_, err = k.client.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		log.Printf("Error encountered while reading secret: %s", err)
	}
	return err
}

func (k KubernetesClientDriver) updateSecret(object interface{}, namespace string, name string) error {
	var err error
	secret := makeSecret(object, namespace, name)
	_, err = k.client.CoreV1().Secrets(namespace).Update(context.TODO(), secret, metav1.UpdateOptions{})
	if err != nil {
		log.Printf("Error encountered while reading secret: %s", err)
	}
	return err
}

func (k KubernetesClientDriver) readSecret(namespace string, name string) *corev1.Secret {
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

func (k KubernetesClientDriver) deleteSecret(namespace string, name string) error {
	err := k.client.CoreV1().Secrets(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		log.Printf("Error encountered while deleting secret: %s", err)
	}
	return err
}

func (k KubernetesClientDriver) CheckAndCreateNamespace(namespace string) (string, error) {
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
		StringData: map[string]string{SecretsKeyName: string(objectBytes)},
		Type:       "Opaque",
	}
}
