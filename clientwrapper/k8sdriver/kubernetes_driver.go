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
	JobType          string = "Job"
	ConfigMapType    string = "ConfigMap"
	CronJobType      string = "CronJob"
	PodType          string = "Pod"
	ServiceType      string = "Service"
	ReplicaSetType   string = "ReplicaSet"
)

type KubernetesClientGetRestricted interface {
	GetJob(name string, namespace string) (batchv1.JobStatus, metav1.LabelSelector, int32)
	Delete(resourceName string, resourceNamespace string, gvk schema.GroupVersionKind) bool
	GetLogs(podNamespace string, selector metav1.LabelSelector) (string, error)
}

type KubernetesClientNamespaceRestricted interface {
	CheckAndCreateNamespace(namespace string) (string, error)
}

type KubernetesClient interface {
	//TODO: Add parameters for Deploy
	Deploy(configPayload *string) (bool, string, string)
	CreateAndDeploy(kind string, objName string, namespace string, imageName string, command []string, args []string, existingConfig string, volumeFilename string, volumeConfig string, variables map[string]string) (bool, string, string)
	//TODO: Add parameters for Delete
	Delete(resourceName string, resourceNamespace string, gvk schema.GroupVersionKind) bool
	DeleteBasedOnConfig(configPayload *string) bool
	CheckAndCreateNamespace(namespace string) (string, error)
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

func (k KubernetesClientDriver) Deploy(configPayload *string) (bool, string, string) {
	//TODO: This method currently expects only one object per file. Add in support for separate YAML files combined into one. This can be done by splitting on the "---" string.
	obj, groupVersionKind, err := getResourceObjectFromYAML(configPayload)
	if err != nil {
		return false, "", ""
	}
	var resourceName string
	var namespace string
	switch obj.(type) {
	//TODO: Types like Pod, Deployment, StatefulSet, etc are missing. They would be almost an exact copy of the ReplicaSet case. Namespaces also need to be added in.
	//TODO: This current flow uses Create, just because it is simpler. Should eventually transition to Apply, which will cover more error cases.
	case *batchv1.Job:
		strongTypeObject := obj.(*batchv1.Job)
		log.Printf("%s matched Job. Deploying...\n", strongTypeObject.Name)
		resourceName = strongTypeObject.Name
		namespace, err = k.CheckAndCreateNamespace(strongTypeObject.Namespace)
		if err != nil {
			log.Printf("The namespace could not be created. Error was %s\n", err)
			return false, "", ""
		}
		_, err = k.client.BatchV1().Jobs(namespace).Create(context.TODO(), strongTypeObject, metav1.CreateOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				return k.deleteAndDeploy(configPayload)
			}
			log.Printf("The deploy step threw an error. Error was %s\n", err)
			return false, "", ""
		}
	case *corev1.Service:
		strongTypeObject := obj.(*corev1.Service)
		log.Printf("%s matched ReplicaSet. Deploying...\n", strongTypeObject.Name)
		resourceName = strongTypeObject.Name
		namespace, err = k.CheckAndCreateNamespace(strongTypeObject.Namespace)
		if err != nil {
			log.Printf("The namespace could not be created. Error was %s\n", err)
			return false, "", ""
		}
		_, err = k.client.CoreV1().Services(namespace).Create(context.TODO(), strongTypeObject, metav1.CreateOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				return k.deleteAndDeploy(configPayload)
			}
			log.Printf("The deploy step threw an error. Error was %s\n", err)
			return false, "", ""
		}
	case *appsv1.ReplicaSet:
		strongTypeObject := obj.(*appsv1.ReplicaSet)
		log.Printf("%s matched ReplicaSet. Deploying...\n", strongTypeObject.Name)
		resourceName = strongTypeObject.Name
		namespace, err = k.CheckAndCreateNamespace(strongTypeObject.Namespace)
		if err != nil {
			log.Printf("The namespace could not be created. Error was %s\n", err)
			return false, "", ""
		}
		_, err = k.client.AppsV1().ReplicaSets(namespace).Create(context.TODO(), strongTypeObject, metav1.CreateOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				return k.deleteAndDeploy(configPayload)
			}
			log.Printf("The deploy step threw an error. Error was %s\n", err)
			return false, "", ""
		}
	default:
		obj, groupVersionKind, err = getUnstructuredResourceObjectFromYAML(configPayload)
		if err != nil {
			return false, "", ""
		}
		strongTypeObject := obj.(*unstructured.Unstructured)
		log.Printf("YAML file matched Unstructured of kind %s. Deploying...\n", groupVersionKind.Kind)
		resourceName = strongTypeObject.GetName()
		namespace, err = k.CheckAndCreateNamespace(strongTypeObject.GetNamespace())
		if err != nil {
			log.Printf("The namespace could not be created. Error was %s\n", err)
			return false, "", ""
		}
		_, err = k.dynamicClient.Resource(schema.GroupVersionResource{
			Group:    groupVersionKind.Group,
			Version:  groupVersionKind.Version,
			Resource: getPluralResourceNameFromKind(groupVersionKind.Kind),
		}).Namespace(namespace).Create(context.TODO(), strongTypeObject, metav1.CreateOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				return k.deleteAndDeploy(configPayload)
			}
			log.Printf("The deploy step threw an error. Error was %s\n", err)
			return false, "", ""
		}
	}
	return true, resourceName, namespace
}

//CreateAndDeploy should only be used for ephemeral runs ONLY. "Kind" should just be discerning between a chron job or a job.
//They are both of type "batch/v1".
func (k KubernetesClientDriver) CreateAndDeploy(kind string, objName string, namespace string, imageName string, command []string, args []string, existingConfig string, volumeFilename string, volumeConfig string, variables map[string]string) (bool, string, string) {
	if imageName == "" {
		imageName = DefaultContainer
	}
	if namespace == "" {
		namespace = corev1.NamespaceDefault
	}
	var configPayload string

	envVars := make([]corev1.EnvVar, 0)
	for key, value := range variables {
		envVars = append(envVars, corev1.EnvVar{Name: key, Value: value})
	}
	if existingConfig != "" {
		obj, _, err := getResourceObjectFromYAML(&existingConfig)
		if err != nil {
			return false, "", ""
		}
		switch obj.(type) {
		case *batchv1.Job:
			strongTypeObject := obj.(*batchv1.Job)
			for idx, val := range strongTypeObject.Spec.Template.Spec.Containers {
				for envidx, _ := range envVars {
					strongTypeObject.Spec.Template.Spec.Containers[idx].Env = append(val.Env, envVars[envidx])
				}
			}
			var data []byte
			data, err = json.Marshal(strongTypeObject)
			if err != nil {
				log.Printf("Error marshalling Job: %s", err)
				return false, "", ""
			}
			configPayload = string(data)
		default:
			log.Printf("Generation only works with Jobs for now.\n")
			return false, "", ""
		}
	} else {
		volumeName := objName + "volume"
		configMapName := objName + "config-map"
		filename := volumeFilename

		if volumeConfig != "" {
			configMapSpec := k.createConfigMap(configMapName, namespace, map[string]string{filename: volumeConfig})
			configMapData, err := json.Marshal(configMapSpec)
			if err != nil {
				log.Printf("Error marshalling ConfigMap: %s", err)
				return false, "", ""
			}
			configMapDataString := string(configMapData)
			success, resourceName, resourceNamespace := k.Deploy(&configMapDataString)
			if !success {
				log.Printf("Error deploying ConfigMap")
				return success, resourceName, resourceNamespace
			}
		}
		var containerSpec = corev1.Container{
			Name:    objName,
			Image:   imageName,
			Command: command,
			Args:    args,
			//WorkingDir: "",
			Ports:   nil,
			EnvFrom: nil,
			Env:     envVars,
		}
		if volumeConfig != "" {
			containerSpec.VolumeMounts = []corev1.VolumeMount{
				{
					Name:      volumeName,
					ReadOnly:  false,
					MountPath: filename,
					SubPath:   filename,
				},
			}
		}
		volumeSpec := corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: configMapName},
					//This pre-configures the file as executable
					DefaultMode:          utilpointer.Int32Ptr(0777),
				},
			},
		}
		if kind == JobType {
			jobSpec := batchv1.Job{
				TypeMeta:   metav1.TypeMeta{Kind: JobType, APIVersion: batchv1.SchemeGroupVersion.String()},
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
			if volumeConfig != "" {
				jobSpec.Spec.Template.Spec.Volumes = []corev1.Volume{volumeSpec}
			}
			data, err := json.Marshal(jobSpec)
			if err != nil {
				log.Printf("Error marshalling Job: %s", err)
				return false, "", ""
			}
			configPayload = string(data)
		} else if kind == CronJobType {
			//TODO: Implement me
		}
	}
	return k.Deploy(&configPayload)
}

func (k KubernetesClientDriver) deleteAndDeploy(configPayload *string) (bool, string, string) {
	success := k.DeleteBasedOnConfig(configPayload)
	if success {
		return k.Deploy(configPayload)
	}
	log.Printf("Failed to delete the existing resource successfully. Aborting deploy.")
	return false, "", ""
}

func (k KubernetesClientDriver) Delete(resourceName string, resourceNamespace string, gvk schema.GroupVersionKind) bool {
	deletionPropogationPolicy := metav1.DeletePropagationBackground
	switch gvk.Kind {
	case JobType:
		err := k.client.BatchV1().Jobs(resourceNamespace).Delete(context.TODO(), resourceName, metav1.DeleteOptions{PropagationPolicy: &deletionPropogationPolicy})
		if err != nil {
			if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "doesn't exist") || strings.Contains(err.Error(), "not found") {
				return true
			}
			log.Printf("The delete step threw an error. Error was %s\n", err)
			return false
		}
	case ServiceType:
		err := k.client.CoreV1().Services(resourceNamespace).Delete(context.TODO(), resourceName, metav1.DeleteOptions{PropagationPolicy: &deletionPropogationPolicy})
		if err != nil {
			if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "doesn't exist") || strings.Contains(err.Error(), "not found") {
				return true
			}
			log.Printf("The delete step threw an error. Error was %s\n", err)
			return false
		}
	case ReplicaSetType:
		err := k.client.AppsV1().ReplicaSets(resourceNamespace).Delete(context.TODO(), resourceName, metav1.DeleteOptions{PropagationPolicy: &deletionPropogationPolicy})
		if err != nil {
			if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "doesn't exist") || strings.Contains(err.Error(), "not found") {
				return true
			}
			log.Printf("The delete step threw an error. Error was %s\n", err)
			return false
		}
	default:
		//GVK may not be fully populated. If it isn't this will fail.
		log.Printf("Default case. Unstructured of kind %s. Deleting...\n", gvk.Kind)
		namespace, _ := k.CheckAndCreateNamespace(resourceNamespace)
		err := k.dynamicClient.Resource(schema.GroupVersionResource{
			Group:    gvk.Group,
			Version:  gvk.Version,
			Resource: getPluralResourceNameFromKind(gvk.Kind),
		}).Namespace(namespace).Delete(context.TODO(), resourceName, metav1.DeleteOptions{PropagationPolicy: &deletionPropogationPolicy})
		if err != nil {
			if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "doesn't exist") || strings.Contains(err.Error(), "not found") {
				return true
			}
			log.Printf("The delete step threw an error. Error was %s\n", err)
			return false
		}
	}
	return true
}

func (k KubernetesClientDriver) DeleteBasedOnConfig(configPayload *string) bool {
	//TODO: This method currently expects only one object per file. Add in support for separate YAML files combined into one. This can be done by splitting on the "---" string.
	deletionPropogationPolicy := metav1.DeletePropagationBackground
	obj, groupVersionKind, err := getUnstructuredResourceObjectFromYAML(configPayload)
	if err != nil {
		return false
	}
	strongTypeObject := obj
	log.Printf("YAML file matched Unstructured of kind %s. Deleting...\n", groupVersionKind.Kind)
	namespace, _ := k.CheckAndCreateNamespace(strongTypeObject.GetNamespace())
	err = k.dynamicClient.Resource(schema.GroupVersionResource{
		Group:    groupVersionKind.Group,
		Version:  groupVersionKind.Version,
		Resource: getPluralResourceNameFromKind(groupVersionKind.Kind),
	}).Namespace(namespace).Delete(context.TODO(), strongTypeObject.GetName(), metav1.DeleteOptions{PropagationPolicy: &deletionPropogationPolicy})
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "doesn't exist") || strings.Contains(err.Error(), "not found") {
			return true
		}
		log.Printf("The delete step threw an error. Error was %s\n", err)
		return false
	}
	return true
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
	return namespace, err
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

func getResourceObjectFromYAML(configPayload *string) (runtime.Object, *schema.GroupVersionKind, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, groupVersionKind, err := decode([]byte(*configPayload), nil, nil)
	if err != nil {
		log.Printf("Error while decoding YAML object. Error was: %s. Now trying to decode as unstructured object...\n", err)
		obj = &unstructured.Unstructured{}
		unstructuredDecoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		_, groupVersionKind, err = unstructuredDecoder.Decode([]byte(*configPayload), nil, obj)
		if err != nil {
			log.Printf("Error while decoding unstructured YAML object. Error was: %s\n", err)
			return nil, nil, err
		}
	}
	return obj, groupVersionKind, err
}

func getUnstructuredResourceObjectFromYAML(configPayload *string) (*unstructured.Unstructured, *schema.GroupVersionKind, error) {
	obj := &unstructured.Unstructured{}
	unstructuredDecoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, groupVersionKind, err := unstructuredDecoder.Decode([]byte(*configPayload), nil, obj)
	if err != nil {
		log.Printf("Error while decoding unstructured YAML object. Error was: %s\n", err)
		return nil, nil, err
	}
	return obj, groupVersionKind, err
}

func getPluralResourceNameFromKind(kind string) string {
	return strings.ToLower(kind) + "s"
}
