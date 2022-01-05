package kubernetesclient

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const (
	secretsKeyName string = "data"
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
	FetchSecretData(name string, namespace string) map[string][]byte
	WatchSecretData(name string, namespace string, handler WatchSecretHandler)
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

func (k KubernetesClientDriver) FetchSecretData(name string, namespace string) map[string][]byte {
	log.Printf("in WatchSecretData, name = %s, namespace = %s\n", name, namespace)
	secret := k.readSecret(namespace, name)
	log.Printf("in WatchSecretData, secret = %v\n", secret)
	if secret != nil {
		log.Printf("in WatchSecretData, secret.Data = %v\n", secret.Data)
		return secret.Data
	}
	return nil
}

func (k KubernetesClientDriver) WatchSecretData(name string, namespace string, handler WatchSecretHandler) {
	factory := informers.NewSharedInformerFactoryWithOptions(k.client, 0, informers.WithNamespace(namespace))
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
	informer.SetWatchErrorHandler(func(r *cache.Reflector, err error) {
		log.Println("failed to watch secret values: ", err)
	})
	// TODO: add channel from context and close it from outside
	informer.Run(make(chan struct{}))
}

func handleSecretInformerEvent(obj interface{}, name string, t SecretChangeType, handler WatchSecretHandler) {
	secret, ok := obj.(*corev1.Secret)
	if !ok {
		log.Println("failed to parse secret data in secret informer")
		return
	}
	if secret.Name != name {
		return
	}
	log.Printf("in WatchSecretData, informer handler. type = %d secret data = %v\n", t, secret.Data)
	handler(t, secret)
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

func makeSecret(object interface{}, namespace string, name string) (*corev1.Secret, error) {
	objectBytes, err := json.Marshal(object)
	if err != nil {
		return nil, err
	}
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
	}, nil
}
