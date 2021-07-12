package k8sdriver

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"log"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilpointer "k8s.io/utils/pointer"
)

const (
	DefaultContainer string = "alpine:3.14"
	JobKindName      string = "Job"
	CronJobKindName  string = "CronJob"
)

type KubernetesClientGetRestricted interface {
	GetJob(name string, namespace string) (batchv1.JobStatus, metav1.LabelSelector, int32)
	GetLogs(podNamespace string, selector metav1.LabelSelector) (string, error)
}

type KubernetesClientNamespaceRestricted interface {
	CheckAndCreateNamespace(namespace string) error
}

type KubernetesClient interface {
	//TODO: Add parameters for Deploy
	Deploy(configPayload string) (bool, string)
	CreateAndDeploy(kind string, objName string, namespace string, imageName string, command []string, args []string, variables map[string]string) (bool, string)
	//TODO: Add parameters for Delete
	Delete(configPayload string) bool
	CheckAndCreateNamespace(namespace string) error
	GetLogs(podNamespace string, selector metav1.LabelSelector) (string, error)
	GetJob(name string, namespace string) (batchv1.JobStatus, metav1.LabelSelector, int32)
	GetSecret(name string, namespace string) map[string][]byte
	//TODO: Update parameters & return type for CheckStatus
	CheckHealthy() bool
	//TODO: Update parameters for ExecInPod
	ExecInPod() bool
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

func (k KubernetesClientDriver) Deploy(configPayload string) (bool, string) {
	//TODO: This method currently expects only one object per file. Add in support for separate YAML files combined into one. This can be done by splitting on the "---" string.
	obj, groupVersionKind, err := getResourceObjectFromYAML(configPayload)
	if err != nil {
		return false, ""
	}
	var namespace string
	switch obj.(type) {
	//TODO: Types like Pod, Deployment, StatefulSet, etc are missing. They would be almost an exact copy of the ReplicaSet case. Namespaces also need to be added in.
	//TODO: This current flow uses Create, just because it is simpler. Should eventually transition to Apply, which will cover more error cases.
	case *batchv1.Job:
		strongTypeObject := obj.(*batchv1.Job)
		log.Printf("%s matched Job. Deploying...\n", strongTypeObject.Name)
		namespace = strongTypeObject.Namespace
		err = k.CheckAndCreateNamespace(namespace)
		if err != nil {
			log.Printf("The namespace could not be created. Error was %s\n", err)
			return false, ""
		}
		_, err = k.client.BatchV1().Jobs(namespace).Create(context.TODO(), strongTypeObject, metav1.CreateOptions{})
		if err != nil {
			log.Printf("The deploy step threw an error. Error was %s\n", err)
			return false, ""
		}
	case *corev1.Service:
		strongTypeObject := obj.(*corev1.Service)
		log.Printf("%s matched ReplicaSet. Deploying...\n", strongTypeObject.Name)
		namespace = strongTypeObject.Namespace
		err = k.CheckAndCreateNamespace(namespace)
		if err != nil {
			log.Printf("The namespace could not be created. Error was %s\n", err)
			return false, ""
		}
		_, err = k.client.CoreV1().Services(namespace).Create(context.TODO(), strongTypeObject, metav1.CreateOptions{})
		if err != nil {
			log.Printf("The deploy step threw an error. Error was %s\n", err)
			return false, ""
		}
	case *appsv1.ReplicaSet:
		strongTypeObject := obj.(*appsv1.ReplicaSet)
		log.Printf("%s matched ReplicaSet. Deploying...\n", strongTypeObject.Name)
		namespace = strongTypeObject.Namespace
		err = k.CheckAndCreateNamespace(namespace)
		if err != nil {
			log.Printf("The namespace could not be created. Error was %s\n", err)
			return false, ""
		}
		_, err = k.client.AppsV1().ReplicaSets(namespace).Create(context.TODO(), strongTypeObject, metav1.CreateOptions{})
		if err != nil {
			log.Printf("The deploy step threw an error. Error was %s\n", err)
			return false, ""
		}
	case *unstructured.Unstructured:
		strongTypeObject := obj.(*unstructured.Unstructured)
		log.Printf("YAML file matched Unstructured of kind %s. Deploying...\n", groupVersionKind.Kind)
		namespace = strongTypeObject.GetNamespace()
		if namespace == "" {
			namespace = corev1.NamespaceDefault
		}
		err = k.CheckAndCreateNamespace(namespace)
		if err != nil {
			log.Printf("The namespace could not be created. Error was %s\n", err)
			return false, ""
		}
		_, err = k.dynamicClient.Resource(schema.GroupVersionResource{
			Group:    groupVersionKind.Group,
			Version:  groupVersionKind.Version,
			Resource: getPluralResourceNameFromKind(groupVersionKind.Kind),
		}).Namespace(namespace).Create(context.TODO(), strongTypeObject, metav1.CreateOptions{})
		if err != nil {
			log.Printf("The deploy step threw an error. Error was %s\n", err)
			return false, ""
		}
	default:
		log.Printf("There was no matching type for the input.\n")
		return false, ""
	}
	return true, namespace
}

//CreateAndDeploy should only be used for ephemeral runs ONLY. "Kind" should just be discerning between a chron job or a job.
//They are both of type "batch/v1".
func (k KubernetesClientDriver) CreateAndDeploy(kind string, objName string, namespace string, imageName string, command []string, args []string, variables map[string]string) (bool, string) {
	if imageName == "" {
		imageName = DefaultContainer
	}
	if namespace == "" {
		namespace = corev1.NamespaceDefault
	}
	envVars := make([]corev1.EnvVar, 0)
	for key, value := range variables {
		envVars = append(envVars, corev1.EnvVar{Name: key, Value: value})
	}
	var containerSpec = corev1.Container{
		Name:       objName,
		Image:      imageName,
		Command:    command,
		Args:       args,
		WorkingDir: "",
		Ports:      nil,
		EnvFrom:    nil,
		Env:        envVars,
		//There are a lot more parameters, probably won't need them in the near future
	}
	var configPayload string
	if kind == JobKindName {
		jobSpec := batchv1.Job{
			TypeMeta:   metav1.TypeMeta{Kind: "Job", APIVersion: batchv1.SchemeGroupVersion.String()},
			ObjectMeta: metav1.ObjectMeta{Name: objName, Namespace: namespace},
			Spec: batchv1.JobSpec{
				BackoffLimit: utilpointer.Int32Ptr(1),
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers:    []corev1.Container{containerSpec},
						RestartPolicy: corev1.RestartPolicyNever,
					},
				},
			},
			Status: batchv1.JobStatus{},
		}
		data, err := json.Marshal(jobSpec)
		if err != nil {
			log.Printf("Error marshalling Job: %s", err)
			return false, ""
		}
		configPayload = string(data)
	} else if kind == CronJobKindName {
		//TODO: Implement me
	}
	return k.Deploy(configPayload)
}

func (k KubernetesClientDriver) Delete(configPayload string) bool {
	//TODO: This method currently expects only one object per file. Add in support for separate YAML files combined into one. This can be done by splitting on the "---" string.
	obj, groupVersionKind, err := getResourceObjectFromYAML(configPayload)
	if err != nil {
		return false
	}
	switch obj.(type) {
	//TODO: Types like Pod, Deployment, StatefulSet, etc are missing. They would be almost an exact copy of the ReplicaSet case. Namespaces also need to be added in.
	//TODO: This current flow uses Create, just because it is simpler. Should eventually transition to Apply, which will cover more error cases.
	case *batchv1.Job:
		strongTypeObject := obj.(*batchv1.Job)
		log.Printf("%s matched ReplicaSet. Deploying...\n", strongTypeObject.Name)
		namespace := strongTypeObject.Namespace
		err = k.client.BatchV1().Jobs(namespace).Delete(context.TODO(), strongTypeObject.GetName(), metav1.DeleteOptions{})
		if err != nil {
			log.Printf("The deploy step threw an error. Error was %s\n", err)
			return false
		}
	case *corev1.Service:
		strongTypeObject := obj.(*corev1.Service)
		log.Printf("%s matched ReplicaSet. Deleting...\n", strongTypeObject.Name)
		namespace := strongTypeObject.Namespace
		err = k.client.CoreV1().Services(namespace).Delete(context.TODO(), strongTypeObject.GetName(), metav1.DeleteOptions{})
		if err != nil {
			log.Printf("The delete step threw an error. Error was %s\n", err)
			return false
		}
	case *appsv1.ReplicaSet:
		strongTypeObject := obj.(*appsv1.ReplicaSet)
		log.Printf("%s matched ReplicaSet. Deleting...\n", strongTypeObject.Name)
		namespace := strongTypeObject.Namespace
		err = k.client.AppsV1().ReplicaSets(namespace).Delete(context.TODO(), strongTypeObject.GetName(), metav1.DeleteOptions{})
		if err != nil {
			log.Printf("The delete step threw an error. Error was %s\n", err)
			return false
		}
	case *unstructured.Unstructured:
		strongTypeObject := obj.(*unstructured.Unstructured)
		log.Printf("YAML file matched Unstructured of kind %s. Deleting...\n", groupVersionKind.Kind)
		namespace := strongTypeObject.GetNamespace()
		if namespace == "" {
			namespace = corev1.NamespaceDefault
		}
		err = k.dynamicClient.Resource(schema.GroupVersionResource{
			Group:    groupVersionKind.Group,
			Version:  groupVersionKind.Version,
			Resource: getPluralResourceNameFromKind(groupVersionKind.Kind),
		}).Namespace(namespace).Delete(context.TODO(), strongTypeObject.GetName(), metav1.DeleteOptions{})
		if err != nil {
			log.Printf("The delete step threw an error. Error was %s\n", err)
			return false
		}
	default:
		log.Printf("There was no matching type for the input.\n")
		return false
	}
	return true
}

func (k KubernetesClientDriver) CheckAndCreateNamespace(namespace string) error {
	if _, err := k.client.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{}); err == nil {
		return nil
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
	return err
}

type GetLogs func(podName string, podNamespace string, selector metav1.LabelSelector)

//Right now only works for simple one container/pod instances created by this driver
func (k KubernetesClientDriver) GetLogs(podNamespace string, selector metav1.LabelSelector) (string, error) {
	labelMap, _ := metav1.LabelSelectorAsMap(&selector)
	podList, err := k.client.CoreV1().Pods(podNamespace).List(context.TODO(), metav1.ListOptions{LabelSelector: labels.SelectorFromSet(labelMap).String()})
	if err != nil {
		return "", err
	}
	var completeLogs []string
	var podLogs string
	for _, pod := range podList.Items {
		completeLogs = append(completeLogs, "Pod name: "+pod.Name)
		podLogs, err = k.getLogsFromSinglePod(pod.Name, pod.Namespace)
		if err != nil {
			return "", err
		}
		completeLogs = append(completeLogs, podLogs)
	}
	return strings.Join(completeLogs, "\n----------\n"), nil
}

func (k KubernetesClientDriver) getLogsFromSinglePod(podName string, podNamespace string) (string, error) {
	podLogOpts := corev1.PodLogOptions{}
	req := k.client.CoreV1().Pods(podNamespace).GetLogs(podName, &podLogOpts)
	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		return "", err
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", err
	}
	str := buf.String()

	return str, nil
}

func (k KubernetesClientDriver) GetJob(name string, namespace string) (batchv1.JobStatus, metav1.LabelSelector, int32) {
	job, err := k.client.BatchV1().Jobs(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		log.Printf("Error getting Job %s", err)
		//This is temporary
		return batchv1.JobStatus{}, metav1.LabelSelector{}, -1
	}
	return job.Status, *job.Spec.Selector, *job.Spec.Parallelism
}

//TODO: This should probably be agnostic to type. Challenge is the Workflow Orchestrator can't be expected to know what the type is.
func (k KubernetesClientDriver) GetSecret(name string, namespace string) map[string][]byte {
	secret, err := k.client.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		log.Printf("Error %s", err)
		return nil
	}
	return secret.Data
}

func (k KubernetesClientDriver) CheckHealthy() bool {
	panic("implement me")
}

func (k KubernetesClientDriver) ExecInPod() bool {
	panic("implement me")
}

func getResourceObjectFromYAML(configPayload string) (runtime.Object, *schema.GroupVersionKind, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, groupVersionKind, err := decode([]byte(configPayload), nil, nil)
	if err != nil {
		log.Printf("Error while decoding YAML object. Error was: %s. Now trying to decode as unstructured object...\n", err)
		obj = &unstructured.Unstructured{}
		unstructuredDecoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		_, groupVersionKind, err = unstructuredDecoder.Decode([]byte(configPayload), nil, obj)
		if err != nil {
			log.Printf("Error while decoding unstructured YAML object. Error was: %s\n", err)
			return nil, nil, err
		}
	}
	return obj, groupVersionKind, err
}

func getPluralResourceNameFromKind(kind string) string {
	return strings.ToLower(kind) + "s"
}

//This is just for testing
//func main() {
//	kubernetesClient := New()
//	//s2 := "apiVersion: apps/v1\nkind: ReplicaSet\nmetadata:\n  name: frontend\n  labels:\n    app: guestbook\n    tier: frontend\nspec:\n  # modify replicas according to your case\n  replicas: 3\n  selector:\n    matchLabels:\n      tier: frontend\n  template:\n    metadata:\n      labels:\n        tier: frontend\n    spec:\n      containers:\n      - name: php-redis\n        image: gcr.io/google_samples/gb-frontend:v3"
//	//s2 := "apiVersion: v1\nkind: Service\nmetadata:\n  name: my-service\nspec:\n  selector:\n    app: MyApp\n  ports:\n    - protocol: TCP\n      port: 80\n      targetPort: 9376"
//	//kubernetesClient.Deploy(s2)
//	s1 := "apiVersion: networking.istio.io/v1alpha3\nkind: VirtualService\nmetadata:\n  name: reviews\n  namespace: guestbook\nspec:\n  hosts:\n    - my-service\n  http:\n    -\n      match:\n        -\n          headers:\n            end-user:\n              exact: jason\n      route:\n        -\n          destination:\n            host: my-service\n            subset: all"
//	kubernetesClient.Deploy(s1)
//}
