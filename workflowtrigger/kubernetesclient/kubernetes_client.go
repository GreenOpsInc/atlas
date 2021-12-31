package kubernetesclient

import (
	"context"
	"log"
	"strings"
	"time"

	"k8s.io/client-go/tools/cache"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/client-go/dynamic/dynamicinformer"

	"greenops.io/workflowtrigger/util/git"
	"greenops.io/workflowtrigger/util/serializer"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	secretsKeyName   string = "data"
	gitCredNamespace string = "gitcred"
)

type KubernetesClient interface {
	StoreGitCred(gitCred git.GitCred, name string) bool
	FetchGitCred(name string) git.GitCred
	FetchSecretData(name string, namespace string) map[string][]byte
	WatchSecretData(name string, namespace string, handler func(action SecretChangeType, obj interface{}))
	StoreTLSCert(cert string, name string, namespace string) bool
}

type KubernetesClientDriver struct {
	client        *kubernetes.Clientset
	dynamicClient dynamic.Interface
}

type SecretChangeType int8

const (
	SecretChangeTypeAdd    SecretChangeType = 1
	SecretChangeTypeUpdate SecretChangeType = 2
	SecretChangeTypeDelete SecretChangeType = 3
)

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
	err := k.storeSecret(gitCred.(interface{}), gitCredNamespace, name)
	if err != nil {
		return false
	}
	return true
}

func (k KubernetesClientDriver) StoreTLSCert(cert string, name string, namespace string) bool {
	err := k.storeSecret(cert, namespace, name)
	if err != nil {
		return false
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

func (k KubernetesClientDriver) WatchSecretData(name string, namespace string, handler func(action SecretChangeType, obj interface{})) {
	log.Println("in WatchSecretData")
	resource := schema.GroupVersionResource{Group: "core", Version: "v1", Resource: "secrets"}
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(k.dynamicClient, time.Second*3, namespace, nil)
	informer := factory.ForResource(resource).Informer()
	log.Printf("in WatchSecretData, informer = %v\n", informer)

	// TODO: fix the error:
	//		reflector.go:127] pkg/mod/k8s.io/client-go@v0.19.6/tools/cache/reflector.go:156:
	//		Failed to watch *unstructured.Unstructured: failed to list *unstructured.Unstructured: secrets.core is forbidden:
	//		User "system:serviceaccount:default:default" cannot list resource "secrets" in API group "core" in the namespace "default"
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			log.Printf("in WatchSecretData, informer add handler. obj = %v\n", obj)
			// TODO: filter by name
			handler(SecretChangeTypeAdd, obj)
		},
		UpdateFunc: func(_, newObj interface{}) {
			log.Printf("in WatchSecretData, informer update handler. newObj = %v\n", newObj)
			// TODO: filter by name
			handler(SecretChangeTypeUpdate, newObj)
		},
		DeleteFunc: func(obj interface{}) {
			log.Printf("in WatchSecretData, informer delete handler. obj = %v\n", obj)
			// TODO: filter by name
			handler(SecretChangeTypeDelete, obj)
		},
	})

	// TODO: add channel from context and close it from outside
	informer.Run(make(chan struct{}))
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
		StringData: map[string]string{secretsKeyName: string(objectBytes)},
		Type:       "Opaque",
	}
}
