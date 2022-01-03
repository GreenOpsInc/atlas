package kubernetesclient

import (
	"context"
	"log"
	"strings"

	"greenops.io/workflowtrigger/util/git"
	"greenops.io/workflowtrigger/util/serializer"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

const (
	OpaqueSecretType corev1.SecretType = "Opaque"
	TLSSecretType    corev1.SecretType = "kubernetes.io/tls"
)

const (
	AtlasTLSSecretName = "atlas-server-tls"
	TLSSecretCrtName   = "tls.crt"
	TLSSecretKeyName   = "tls.key"
)

type WatchSecretHandler func(action SecretChangeType, secret *corev1.Secret)

type TLSSecretData struct {
	Crt string `json:"crt"`
	Key string `json:"key"`
}

type KubernetesClient interface {
	StoreGitCred(gitCred git.GitCred, name string) bool
	StoreServerTLSConf(cert string, key string, namespace string) bool
	FetchGitCred(name string) git.GitCred
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

func (k KubernetesClientDriver) StoreServerTLSConf(cert string, key string, namespace string) bool {
	existing := k.readSecret(namespace, AtlasTLSSecretName)
	data := &TLSSecretData{cert, key}

	if existing == nil {
		if err := k.createSecret(data, namespace, AtlasTLSSecretName, TLSSecretType); err != nil {
			return false
		}
	} else if string(existing.Data[TLSSecretCrtName]) != cert || string(existing.Data[TLSSecretKeyName]) != key {
		if err := k.updateSecret(data, namespace, AtlasTLSSecretName, TLSSecretType); err != nil {
			return false
		}
	}
	return true
}

func (k KubernetesClientDriver) FetchGitCred(name string) git.GitCred {
	secret := k.readSecret(gitCredNamespace, name)
	if secret != nil {
		if val, ok := secret.StringData[secretsKeyName]; ok {
			return git.UnmarshallGitCredString(val)
		}
		return nil
	}
	return nil
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

func (k KubernetesClientDriver) storeSecret(object interface{}, namespace string, name string) error {
	var err error
	_, err = k.CheckAndCreateNamespace(namespace)
	if err != nil {
		return err
	}
	if object == nil {
		return k.deleteSecret(namespace, name)
	} else if k.readSecret(namespace, name) == nil {
		return k.createSecret(object, namespace, name, OpaqueSecretType)
	}
	return k.updateSecret(object, namespace, name, OpaqueSecretType)
}

func (k KubernetesClientDriver) createSecret(object interface{}, namespace string, name string, t corev1.SecretType) error {
	var err error
	secret := makeSecret(object, namespace, name, t)
	_, err = k.client.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		log.Printf("Error encountered while reading secret: %s", err)
	}
	return err
}

func (k KubernetesClientDriver) updateSecret(object interface{}, namespace string, name string, t corev1.SecretType) error {
	var err error
	secret := makeSecret(object, namespace, name, t)
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

func makeSecret(object interface{}, namespace string, name string, t corev1.SecretType) *corev1.Secret {
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
		Type:       t,
	}
}
