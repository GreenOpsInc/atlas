package kubernetesclient

import (
	"context"
	"greenops.io/workflowtrigger/util/git"
	"greenops.io/workflowtrigger/util/serializer"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"strings"
)

const (
	secretsKeyName   string = "data"
	gitCredNamespace string = "gitcred"
)

type KubernetesClient interface {
	StoreGitCred(gitCred git.GitCred, name string) bool
	FetchGitCred(name string) git.GitCred
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
	err := k.storeSecret(gitCred.(interface{}), gitCredNamespace, name)
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
