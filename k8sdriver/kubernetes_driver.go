package main

import (
	"context"
	"log"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KubernetesClient interface {
	//TODO: Add parameters for Deploy
	Deploy(configPayload string) bool
	//TODO: Add parameters for Delete
	Delete() bool
	//TODO: Update parameters & return type for CheckStatus
	CheckStatus() bool
	//TODO: Update parameters for ExecInPod
	ExecInPod() bool
}

type KubernetesClientDriver struct {
	client        *kubernetes.Clientset
	dynamicClient dynamic.Interface
}

//var kindResourceMap map[string]string
var kindResourceMap = map[string]string{
	//TODO: This list is obviously not complete. As we do more testing with CRDs, more items should be added to the list.
	"VirtualService": "virtualservices",
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

func (k KubernetesClientDriver) Deploy(configPayload string) bool {
	//TODO: This method currently expects only one object per file. Add in support for separate YAML files combined into one. This can be done by splitting on the "---" string.
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, groupVersionKind, err := decode([]byte(configPayload), nil, nil)
	if err != nil {
		log.Printf("Error while decoding YAML object. Error was: %s. Now trying to decode as unstructured object...\n", err)
		obj = &unstructured.Unstructured{}
		unstructuredDecoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		_, groupVersionKind, err = unstructuredDecoder.Decode([]byte(configPayload), nil, obj)
		if err != nil {
			log.Printf("Error while decoding unstructured YAML object. Error was: %s\n", err)
			return false
		}
	}
	switch obj.(type) {
	//TODO: Types like Pod, Deployment, StatefulSet, etc are missing. They would be almost an exact copy of the ReplicaSet case. Namespaces also need to be added in.
	//TODO: This current flow uses Create, just because it is simpler. Should eventually transition to Apply, which will cover more error cases.
	case *corev1.Service:
		strongTypeObject := obj.(*corev1.Service)
		log.Printf("%s matched ReplicaSet. Deploying...\n", strongTypeObject.Name)
		_, err = k.client.CoreV1().Services(corev1.NamespaceDefault).Create(context.TODO(), strongTypeObject, metav1.CreateOptions{})
		if err != nil {
			log.Printf("The deploy step threw an error. Error was %s\n", err)
			return false
		}
	case *appsv1.ReplicaSet:
		strongTypeObject := obj.(*appsv1.ReplicaSet)
		log.Printf("%s matched ReplicaSet. Deploying...\n", strongTypeObject.Name)
		_, err = k.client.AppsV1().ReplicaSets(corev1.NamespaceDefault).Create(context.TODO(), strongTypeObject, metav1.CreateOptions{})
		if err != nil {
			log.Printf("The deploy step threw an error. Error was %s\n", err)
			return false
		}
	case *unstructured.Unstructured:
		strongTypeObject := obj.(*unstructured.Unstructured)
		log.Printf("YAML file matched Unstructured of kind %s. Deploying...\n", groupVersionKind.Kind)
		resourceName, ok := kindResourceMap[groupVersionKind.Kind]
		if !ok {
			log.Printf("The key %s was not in the kindResourceMap\n", groupVersionKind.Kind)
			return false
		}
		_, err = k.dynamicClient.Resource(schema.GroupVersionResource{
			Group:    groupVersionKind.Group,
			Version:  groupVersionKind.Version,
			Resource: resourceName,
		}).Namespace(corev1.NamespaceDefault).Create(context.TODO(), strongTypeObject, metav1.CreateOptions{})
		if err != nil {
			log.Printf("The deploy step threw an error. Error was %s\n", err)
			return false
		}
	default:
		log.Printf("There was no matching type for the input.\n")
		return false
	}
	return true
}

func (k KubernetesClientDriver) Delete() bool {
	panic("implement me")
}

func (k KubernetesClientDriver) CheckStatus() bool {
	panic("implement me")
}

func (k KubernetesClientDriver) ExecInPod() bool {
	panic("implement me")
}

//This is just for testing
//func main() {
//	kubernetesClient := New()
//	//s2 := "apiVersion: apps/v1\nkind: ReplicaSet\nmetadata:\n  name: frontend\n  labels:\n    app: guestbook\n    tier: frontend\nspec:\n  # modify replicas according to your case\n  replicas: 3\n  selector:\n    matchLabels:\n      tier: frontend\n  template:\n    metadata:\n      labels:\n        tier: frontend\n    spec:\n      containers:\n      - name: php-redis\n        image: gcr.io/google_samples/gb-frontend:v3"
//	//s2 := "apiVersion: v1\nkind: Service\nmetadata:\n  name: my-service\nspec:\n  selector:\n    app: MyApp\n  ports:\n    - protocol: TCP\n      port: 80\n      targetPort: 9376"
//	//kubernetesClient.Deploy(s2)
//	s1 := "apiVersion: networking.istio.io/v1alpha3\nkind: VirtualService\nmetadata:\n  name: reviews\nspec:\n  hosts:\n  - my-service\n  http:\n  - match:\n    - headers:\n        end-user:\n          exact: jason\n    route:\n    - destination:\n        host: my-service\n        subset: all"
//	kubernetesClient.Deploy(s1)
//}
