package k8sdriver

import (
	"bytes"
	"context"
	"encoding/json"
	goerrors "errors"
	"io"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"log"
	"strings"

	"github.com/greenopsinc/util/clientrequest"

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
	DeploymentType   string = "Deployment"
	DaemonSetType    string = "DaemonSet"
	StatefulSetType  string = "StatefulSet"
)

type AggregateResult struct {
	Cluster      string     `json:"cluster"`
	ResourceList []Resource `json:"resourceList"`
}

type Resource struct {
	Kind         string     `json:"kind"`
	Name         string     `json:"name"`
	Version      string     `json:"version"`
	ResourceList []Resource `json:"resource_list"`
}
type patchStringValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

type KubernetesClientGetRestricted interface {
	GetJob(name string, namespace string) (batchv1.JobStatus, metav1.LabelSelector, int32)
	Delete(resourceName string, resourceNamespace string, gvk schema.GroupVersionKind) error
	GetLogs(podNamespace string, selector metav1.LabelSelector) (string, error)
	GetConfigMapNameFromTask(taskName string) string
	GetConfigMap(name string, namespace string) (map[string][]byte, error)
	StoreConfigMap(name string, namespace string, data map[string][]byte) error
}

type KubernetesClientNamespaceSecretRestricted interface {
	CheckAndCreateNamespace(namespace string) (string, error)
	GetSecret(name string, namespace string) map[string][]byte
}

type KubernetesClient interface {
	//TODO: Add parameters for Deploy
	Deploy(configPayload *string) (string, string, error)
	CreateAndDeploy(kind string, objName string, namespace string, imageName string, command []string, args []string, existingConfig string, volumeFilename string, volumeConfig string, variables []corev1.EnvVar) (string, string, error)
	//TODO: Add parameters for Delete
	Delete(resourceName string, resourceNamespace string, gvk schema.GroupVersionKind) error
	DeleteBasedOnConfig(configPayload *string) error
	CheckAndCreateNamespace(namespace string) (string, error)
	GetLogs(podNamespace string, selector metav1.LabelSelector) (string, error)
	GetJob(name string, namespace string) (batchv1.JobStatus, metav1.LabelSelector, int32)
	GetConfigMapNameFromTask(taskName string) string
	GetSecret(name string, namespace string) map[string][]byte
	GetConfigMap(name string, namespace string) (map[string][]byte, error)
	Label(gvkGroup clientrequest.GvkGroupRequest, resourcesLabel string) error
	Aggregate(cluster string, namespace string) (AggregateResult, error)
	DeleteByLabel(label string, namespace string) error
	//TODO: Update parameters & return type for CheckStatus
	CheckHealthy() bool
	//TODO: Update parameters for ExecInPod
	ExecInPod() bool
	StoreConfigMap(name string, namespace string, data map[string][]byte) error
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

func (k KubernetesClientDriver) Deploy(configPayload *string) (string, string, error) {
	//TODO: This method currently expects only one object per file. Add in support for separate YAML files combined into one. This can be done by splitting on the "---" string.
	obj, groupVersionKind, err := getResourceObjectFromYAML(configPayload)
	if err != nil {
		return "", "", err
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
			return "", "", err
		}
		_, err = k.client.BatchV1().Jobs(namespace).Create(context.TODO(), strongTypeObject, metav1.CreateOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				return k.deleteAndDeploy(configPayload)
			}
			log.Printf("The deploy step threw an error. Error was %s\n", err)
			return "", "", err
		}
	case *corev1.Service:
		strongTypeObject := obj.(*corev1.Service)
		log.Printf("%s matched ReplicaSet. Deploying...\n", strongTypeObject.Name)
		resourceName = strongTypeObject.Name
		namespace, err = k.CheckAndCreateNamespace(strongTypeObject.Namespace)
		if err != nil {
			log.Printf("The namespace could not be created. Error was %s\n", err)
			return "", "", err
		}
		_, err = k.client.CoreV1().Services(namespace).Create(context.TODO(), strongTypeObject, metav1.CreateOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				return k.deleteAndDeploy(configPayload)
			}
			log.Printf("The deploy step threw an error. Error was %s\n", err)
			return "", "", err
		}
	case *appsv1.ReplicaSet:
		strongTypeObject := obj.(*appsv1.ReplicaSet)
		log.Printf("%s matched ReplicaSet. Deploying...\n", strongTypeObject.Name)
		resourceName = strongTypeObject.Name
		namespace, err = k.CheckAndCreateNamespace(strongTypeObject.Namespace)
		if err != nil {
			log.Printf("The namespace could not be created. Error was %s\n", err)
			return "", "", err
		}
		_, err = k.client.AppsV1().ReplicaSets(namespace).Create(context.TODO(), strongTypeObject, metav1.CreateOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				return k.deleteAndDeploy(configPayload)
			}
			log.Printf("The deploy step threw an error. Error was %s\n", err)
			return "", "", err
		}
	default:
		obj, groupVersionKind, err = getUnstructuredResourceObjectFromYAML(configPayload)
		if err != nil {
			return "", "", err
		}
		strongTypeObject := obj.(*unstructured.Unstructured)
		log.Printf("YAML file matched Unstructured of kind %s. Deploying...\n", groupVersionKind.Kind)
		resourceName = strongTypeObject.GetName()
		namespace, err = k.CheckAndCreateNamespace(strongTypeObject.GetNamespace())
		if err != nil {
			log.Printf("The namespace could not be created. Error was %s\n", err)
			return "", "", err
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
			return "", "", err
		}
	}
	return resourceName, namespace, nil
}

func (k KubernetesClientDriver) GetVolumeNameFromTask(taskName string) string {
	return taskName + "volume"
}

func (k KubernetesClientDriver) GetConfigMapNameFromTask(taskName string) string {
	return taskName + "config-map"
}

//CreateAndDeploy should only be used for ephemeral runs ONLY. "Kind" should just be discerning between a chron job or a job.
//They are both of type "batch/v1".
func (k KubernetesClientDriver) CreateAndDeploy(kind string, objName string, namespace string, imageName string, command []string, args []string, existingConfig string, volumeFilename string, volumeConfig string, variables []corev1.EnvVar) (string, string, error) {
	if imageName == "" {
		imageName = DefaultContainer
	}
	if namespace == "" {
		namespace = corev1.NamespaceDefault
	}
	var configPayload string

	if existingConfig != "" {
		obj, _, err := getResourceObjectFromYAML(&existingConfig)
		if err != nil {
			return "", "", err
		}
		switch obj.(type) {
		case *batchv1.Job:
			strongTypeObject := obj.(*batchv1.Job)
			for idx, val := range strongTypeObject.Spec.Template.Spec.Containers {
				for envidx, _ := range variables {
					strongTypeObject.Spec.Template.Spec.Containers[idx].Env = append(val.Env, variables[envidx])
				}
			}
			var data []byte
			data, err = json.Marshal(strongTypeObject)
			if err != nil {
				log.Printf("Error marshalling Job: %s", err)
				return "", "", err
			}
			configPayload = string(data)
		default:
			log.Printf("Generation only works with Jobs for now.\n")
			return "", "", goerrors.New("generation only works with Jobs for now")
		}
	} else {
		volumeName := k.GetVolumeNameFromTask(objName)
		configMapName := k.GetConfigMapNameFromTask(objName)
		filename := volumeFilename

		if volumeConfig != "" {
			configMapSpec := k.createConfigMap(configMapName, namespace, map[string]string{filename: volumeConfig})
			configMapData, err := json.Marshal(configMapSpec)
			if err != nil {
				log.Printf("Error marshalling ConfigMap: %s", err)
				return "", "", err
			}
			configMapDataString := string(configMapData)
			resourceName, resourceNamespace, err := k.Deploy(&configMapDataString)
			if err != nil {
				log.Printf("Error deploying ConfigMap")
				return resourceName, resourceNamespace, err
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
			Env:     variables,
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
					DefaultMode: utilpointer.Int32Ptr(0777),
				},
			},
		}
		if kind == JobType {
			jobSpec := batchv1.Job{
				TypeMeta:   metav1.TypeMeta{Kind: JobType, APIVersion: batchv1.SchemeGroupVersion.String()},
				ObjectMeta: metav1.ObjectMeta{Name: objName, Namespace: namespace},
				Spec: batchv1.JobSpec{
					Completions:  utilpointer.Int32Ptr(1),
					BackoffLimit: utilpointer.Int32Ptr(0),
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers:    []corev1.Container{containerSpec},
							RestartPolicy: corev1.RestartPolicyNever,
						},
					},
				},
			}
			if volumeConfig != "" {
				jobSpec.Spec.Template.Spec.Volumes = []corev1.Volume{volumeSpec}
			}
			data, err := json.Marshal(jobSpec)
			if err != nil {
				log.Printf("Error marshalling Job: %s", err)
				return "", "", nil
			}
			configPayload = string(data)
		} else if kind == CronJobType {
			//TODO: Implement me
		}
	}
	return k.Deploy(&configPayload)
}

func (k KubernetesClientDriver) deleteAndDeploy(configPayload *string) (string, string, error) {
	err := k.DeleteBasedOnConfig(configPayload)
	if err == nil {
		return k.Deploy(configPayload)
	}
	log.Printf("Failed to delete the existing resource successfully. Aborting deploy.")
	return "", "", err
}

func (k KubernetesClientDriver) Delete(resourceName string, resourceNamespace string, gvk schema.GroupVersionKind) error {
	deletionPropogationPolicy := metav1.DeletePropagationBackground
	switch gvk.Kind {
	case JobType:
		err := k.client.BatchV1().Jobs(resourceNamespace).Delete(context.TODO(), resourceName, metav1.DeleteOptions{PropagationPolicy: &deletionPropogationPolicy})
		if err != nil {
			if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "doesn't exist") || strings.Contains(err.Error(), "not found") {
				return nil
			}
			log.Printf("The delete step threw an error. Error was %s\n", err)
			return err
		}
	case ServiceType:
		err := k.client.CoreV1().Services(resourceNamespace).Delete(context.TODO(), resourceName, metav1.DeleteOptions{PropagationPolicy: &deletionPropogationPolicy})
		if err != nil {
			if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "doesn't exist") || strings.Contains(err.Error(), "not found") {
				return nil
			}
			log.Printf("The delete step threw an error. Error was %s\n", err)
			return err
		}
	case ReplicaSetType:
		err := k.client.AppsV1().ReplicaSets(resourceNamespace).Delete(context.TODO(), resourceName, metav1.DeleteOptions{PropagationPolicy: &deletionPropogationPolicy})
		if err != nil {
			if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "doesn't exist") || strings.Contains(err.Error(), "not found") {
				return nil
			}
			log.Printf("The delete step threw an error. Error was %s\n", err)
			return err
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
				return nil
			}
			log.Printf("The delete step threw an error. Error was %s\n", err)
			return err
		}
	}
	return nil
}

func (k KubernetesClientDriver) DeleteBasedOnConfig(configPayload *string) error {
	//TODO: This method currently expects only one object per file. Add in support for separate YAML files combined into one. This can be done by splitting on the "---" string.
	deletionPropogationPolicy := metav1.DeletePropagationBackground
	obj, groupVersionKind, err := getUnstructuredResourceObjectFromYAML(configPayload)
	if err != nil {
		return err
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
			return nil
		}
		log.Printf("The delete step threw an error. Error was %s\n", err)
		return err
	}
	return nil
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

func (k KubernetesClientDriver) GetConfigMap(name string, namespace string) (map[string][]byte, error) {
	configMap, err := k.client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return make(map[string][]byte), nil
		}
		log.Printf("Error %s", err)
		return nil, err
	}
	return configMap.BinaryData, nil
}

func (k KubernetesClientDriver) StoreConfigMap(name string, namespace string, data map[string][]byte) error {
	configMap := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       ConfigMapType,
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		BinaryData: data,
	}
	_, err := k.client.CoreV1().ConfigMaps(namespace).Update(context.TODO(), &configMap, metav1.UpdateOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			_, err = k.client.CoreV1().ConfigMaps(namespace).Create(context.TODO(), &configMap, metav1.CreateOptions{})
			if err == nil {
				return nil
			}
			return err
		}
		log.Printf("Error %s", err)
		return err
	}
	return nil
}

func (k KubernetesClientDriver) Label(gvkGroup clientrequest.GvkGroupRequest, resourcesLabel string) error {
	var label_path string
	label_path = "/metadata/labels/" + resourcesLabel
	payload := []patchStringValue{{
		Op:    "replace",
		Path:  label_path,
		Value: "True",
	}}
	payloadBytes, _ := json.Marshal(payload)
	for _, value := range gvkGroup.ResourceList {
		switch value.Kind {
		case PodType:
			_, err := k.client.CoreV1().Pods(value.ResourceNamespace).Patch(context.TODO(), value.ResourceName, types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
			if err != nil {
				if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "doesn't exist") || strings.Contains(err.Error(), "not found") {
					return nil
				}
				log.Printf("The label step threw an error. Error was %s\n", err)
				return err
			}
		case ReplicaSetType:
			_, err := k.client.AppsV1().ReplicaSets(value.ResourceNamespace).Patch(context.TODO(), value.ResourceName, types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
			if err != nil {
				if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "doesn't exist") || strings.Contains(err.Error(), "not found") {
					return nil
				}
				log.Printf("The label step threw an error. Error was %s\n", err)
				return err
			}
			log.Printf("Successfully labeled replica set with label ", resourcesLabel)
		case DaemonSetType:
			_, err := k.client.AppsV1().DaemonSets(value.ResourceNamespace).Patch(context.TODO(), value.ResourceName, types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
			if err != nil {
				if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "doesn't exist") || strings.Contains(err.Error(), "not found") {
					return nil
				}
				log.Printf("The label step threw an error. Error was %s\n", err)
				return err
			}
		case StatefulSetType:
			_, err := k.client.AppsV1().StatefulSets(value.ResourceNamespace).Patch(context.TODO(), value.ResourceName, types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
			if err != nil {
				if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "doesn't exist") || strings.Contains(err.Error(), "not found") {
					return nil
				}
				log.Printf("The label step threw an error. Error was %s\n", err)
				return err
			}
		case DeploymentType:
			_, err := k.client.AppsV1().Deployments(value.ResourceNamespace).Patch(context.TODO(), value.ResourceName, types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
			if err != nil {
				if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "doesn't exist") || strings.Contains(err.Error(), "not found") {
					return nil
				}
				log.Printf("The label step threw an error. Error was %s\n", err)
				return err
			}
			log.Printf("Successfully labeled deployment ", resourcesLabel)
		case ServiceType:
			_, err := k.client.CoreV1().Services(value.ResourceNamespace).Patch(context.TODO(), value.ResourceName, types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
			if err != nil {
				if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "doesn't exist") || strings.Contains(err.Error(), "not found") {
					return nil
				}
				log.Printf("The label step threw an error. Error was %s\n", err)
				return err
			}
		default:
			//GVK may not be fully populated. If it isn't this will fail.
			log.Printf("Default case. Unstructured of kind %s. Labeling...\n", value.Kind)
			namespace, _ := k.CheckAndCreateNamespace(value.ResourceNamespace)
			_, err := k.dynamicClient.Resource(schema.GroupVersionResource{
				Group:    value.Group,
				Version:  value.Version,
				Resource: getPluralResourceNameFromKind(value.Kind),
			}).Namespace(namespace).Patch(context.TODO(), value.ResourceName, types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
			if err != nil {
				if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "doesn't exist") || strings.Contains(err.Error(), "not found") {
					return nil
				}
				log.Printf("The label step threw an error. Error was %s\n", err)
				return err
			}
		}

	}

	return nil
}

func (k KubernetesClientDriver) Aggregate(cluster string, namespace string) (AggregateResult, error) {
	groupNameSuffix := "-group"

	var data AggregateResult
	data.Cluster = cluster

	var atlasGroupsList []Resource

	podsMap := make(map[string]corev1.Pod)
	pods, err := k.client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Printf("The aggregate step threw an error fetching pods. Error was %s\n", err)
		return data, err
	}

	log.Printf("Number of pods: %d", len(pods.Items))

	// Create a mapping from pod name to k8s client pod object
	for _, pod := range pods.Items {
		podsMap[pod.Name] = pod
	}

	replicaSetsMap := make(map[string]appsv1.ReplicaSet)
	replicasets, err := k.client.AppsV1().ReplicaSets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Printf("The aggregate step threw an error fetching replica sets. Error was %s\n", err)
		return data, err
	}
	for _, replicaSet := range replicasets.Items {
		replicaSetsMap[replicaSet.Name] = replicaSet
	}

	daemonSetsMap := make(map[string]appsv1.DaemonSet)
	daemonsets, err := k.client.AppsV1().DaemonSets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Printf("The aggregate step threw an error fetching daemon sets. Error was %s\n", err)
		return data, err
	}
	for _, daemonSet := range daemonsets.Items {
		daemonSetsMap[daemonSet.Name] = daemonSet
	}

	statefulSetsMap := make(map[string]appsv1.StatefulSet)
	statefulSets, err := k.client.AppsV1().StatefulSets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Printf("The aggregate step threw an error fetching stateful sets. Error was %s\n", err)
		return data, err
	}
	for _, statefulSet := range statefulSets.Items {
		statefulSetsMap[statefulSet.Name] = statefulSet
	}

	services, err := k.client.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Printf("The aggregate step threw an error fetching services. Error was %s\n", err)
		return data, err
	}

	log.Printf("Number of services: %d", len(services.Items))

	//deploymentsMap := make(map[string]appsv1.Deployment)
	deployments, err := k.client.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Printf("The aggregate step threw an error fetching deployments. Error was %s\n", err)
		return data, err
	}

	log.Printf("Number of deployments: %d", len(deployments.Items))

	endpoints, err := k.client.CoreV1().Endpoints(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Printf("The aggregate step threw an error fetching endpoints. Error was %s\n", err)
		return data, err
	}

	// This map is used to track which AtlasGroup index a replica set, daemon set, or stateful set belongs to
	indexMap := make(map[string]int)

	// Go through the services and find the corresponding endpoints to get the pods
	for _, service := range services.Items {
		if service.Name == "kubernetes" {
			continue
		}
		log.Printf("Going through %s endpoints", len(endpoints.Items))
		for _, ep := range endpoints.Items {
			if ep.Name == service.Name {
				if len(ep.Subsets) == 0 {
					// Lone service
					data.ResourceList = append(data.ResourceList, Resource{ServiceType, service.Name, "core/v1", []Resource{}})
					continue
				}
				for _, subset := range ep.Subsets {
					addresses := append(append([]corev1.EndpointAddress{}, subset.Addresses...), subset.NotReadyAddresses...)

					for _, addr := range addresses {
						log.Printf("Found addresses for service: %s", service.Name)
						// Means that service is not part of an AtlasGroup
						if addr.TargetRef == nil || addr.TargetRef.Kind != "Pod" {
							data.ResourceList = append(data.ResourceList, Resource{ServiceType, service.Name, "core/v1", []Resource{}})
						} else {
							podName := addr.TargetRef.Name
							log.Printf("Pod for this service is: %s", podName)
							owners := podsMap[podName].OwnerReferences
							log.Printf("Number of owners for pod: %d", len(owners))
							// This case means the pod is part of a service, but not a set or deployment so ignore it :)
							if len(owners) <= 0 {
								continue
							}

							// Get the name of the replica set or daemon set that this pod belongs to
							owner := owners[0]
							log.Printf("Owner for this pod: %s", owner.Name)

							// We have seen this set before, so it already is part of an AtlasGroup
							if _, ok := indexMap[owner.Name]; ok {
								continue
							} else {
								// This set has not been seen, so mark which AtlasGroup it will belong to and add the service to this group
								indexMap[owner.Name] = len(atlasGroupsList)
								log.Printf("Adding service to atlas group at index %d", len(atlasGroupsList))
								atlasGroupsList = append(atlasGroupsList, Resource{"AtlasGroup", service.Name + groupNameSuffix, "N/A", []Resource{{ServiceType, service.Name, "core/v1", []Resource{}}}})
							}

						}
					}
				}
			}
		}
	}

	// Go through pods, find the owning sets and add them to their respective AtlasGroups (create the group if
	// the set doesn't have a service and therefore hasn't been seen yet).

	// Make sure we don't reprocess a set more than once since multiple pods belong to the same set
	marked := make(map[string]bool)

	for _, pod := range podsMap {
		owners := pod.OwnerReferences
		if len(owners) > 0 {
			owner := owners[0]
			setName := owner.Name
			if _, ok := marked[setName]; ok {
				continue
			}
			marked[setName] = true

			// Check if the set has an owning deployment. If it does, and we already have a group for them, add it.
			// Create the AtlasGroup if the set does have an owning deployment, but we have not seen it yet and add both.
			// If the set is not part of any deployment we will cover it later.

			if replicaSet, ok := replicaSetsMap[setName]; ok {
				// Find all owning deployments of this set
				for _, rsOwner := range replicaSet.OwnerReferences {
					if rsOwner.Kind != DeploymentType {
						continue
					}

					replicaSetResource := Resource{ReplicaSetType, setName, owner.APIVersion, []Resource{}}
					deploymentResource := Resource{DeploymentType, rsOwner.Name, rsOwner.APIVersion, []Resource{}}

					if _, in := indexMap[setName]; in {
						atlasGroupsList[indexMap[setName]].ResourceList = append(atlasGroupsList[indexMap[setName]].ResourceList, replicaSetResource)
						atlasGroupsList[indexMap[setName]].ResourceList = append(atlasGroupsList[indexMap[setName]].ResourceList, deploymentResource)
					} else {
						atlasGroupsList = append(atlasGroupsList, Resource{"AtlasGroup", rsOwner.Name + groupNameSuffix, "N/A", []Resource{replicaSetResource, deploymentResource}})
					}
				}
			} else if daemonSet, ok := daemonSetsMap[setName]; ok {
				// Find all owning deployments of this set
				for _, dsOwner := range daemonSet.OwnerReferences {
					if dsOwner.Kind != DeploymentType {
						continue
					}

					daemonSetResource := Resource{DaemonSetType, setName, owner.APIVersion, []Resource{}}
					deploymentResource := Resource{DeploymentType, dsOwner.Name, dsOwner.APIVersion, []Resource{}}

					if _, in := indexMap[setName]; in {
						atlasGroupsList[indexMap[setName]].ResourceList = append(atlasGroupsList[indexMap[setName]].ResourceList, daemonSetResource)
						atlasGroupsList[indexMap[setName]].ResourceList = append(atlasGroupsList[indexMap[setName]].ResourceList, deploymentResource)
					} else {
						atlasGroupsList = append(atlasGroupsList, Resource{"AtlasGroup", dsOwner.Name + groupNameSuffix, "N/A", []Resource{daemonSetResource, deploymentResource}})
					}

				}
			} else if statefulSet, ok := statefulSetsMap[setName]; ok {
				// Find all owning deployments of this set
				for _, ssOwner := range statefulSet.OwnerReferences {
					if ssOwner.Kind != DeploymentType {
						continue
					}
					statefulSetResource := Resource{StatefulSetType, setName, owner.APIVersion, []Resource{}}
					deploymentResource := Resource{DeploymentType, ssOwner.Name, ssOwner.APIVersion, []Resource{}}

					if _, in := indexMap[setName]; in {
						atlasGroupsList[indexMap[setName]].ResourceList = append(atlasGroupsList[indexMap[setName]].ResourceList, statefulSetResource)
						atlasGroupsList[indexMap[setName]].ResourceList = append(atlasGroupsList[indexMap[setName]].ResourceList, deploymentResource)
					} else {
						atlasGroupsList = append(atlasGroupsList, Resource{"AtlasGroup", ssOwner.Name + groupNameSuffix, "N/A", []Resource{statefulSetResource, deploymentResource}})
					}
				}
			}

			//Identify any lone pods and add them to the pods list.
		} else {
			data.ResourceList = append(data.ResourceList, Resource{PodType, pod.Name, "core/v1", []Resource{}})
		}
	}

	// Now add any sets without pods or owning deployments to their respective lists
	for _, replicaSet := range replicaSetsMap {
		if _, ok := indexMap[replicaSet.Name]; ok {
			continue
		} else {
			data.ResourceList = append(data.ResourceList, Resource{ReplicaSetType, replicaSet.Name, "apps/v1", []Resource{}})
		}
	}

	for _, daemonSet := range daemonSetsMap {
		if _, ok := indexMap[daemonSet.Name]; ok {
			continue
		} else {
			data.ResourceList = append(data.ResourceList, Resource{DaemonSetType, daemonSet.Name, "apps/v1", []Resource{}})
		}
	}

	for _, statefulSet := range statefulSetsMap {
		if _, ok := indexMap[statefulSet.Name]; ok {
			continue
		} else {
			data.ResourceList = append(data.ResourceList, Resource{StatefulSetType, statefulSet.Name, "apps/v1", []Resource{}})
		}
	}

	data.ResourceList = append(data.ResourceList, atlasGroupsList...)

	return data, nil
}

func (k KubernetesClientDriver) DeleteByLabel(label string, namespace string) error {
	labelDeployment := labels.SelectorFromSet(labels.Set(map[string]string{label: "True"}))
	listDeploymentOptions := metav1.ListOptions{
		LabelSelector: labelDeployment.String(),
	}
	err := k.client.AppsV1().Deployments(namespace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, listDeploymentOptions)
	if err != nil {
		if !errors.IsNotFound(err) {
			log.Printf("Error when deleting by label. Error was %s\n", err)
			return err
		}
	}

	labelService := labels.SelectorFromSet(labels.Set(map[string]string{label: "True"}))
	listServiceOptions := metav1.ListOptions{
		LabelSelector: labelService.String(),
	}
	services, err := k.client.CoreV1().Services(namespace).List(context.TODO(), listServiceOptions)
	for _, service := range services.Items {
		err = k.client.CoreV1().Services(namespace).Delete(context.TODO(), service.Name, metav1.DeleteOptions{})
		if !errors.IsNotFound(err) {
			log.Printf("Error when deleting by label. Error was %s\n", err)
			return err
		}
	}

	labelReplicaSet := labels.SelectorFromSet(labels.Set(map[string]string{label: "True"}))
	listReplicaSetOptions := metav1.ListOptions{
		LabelSelector: labelReplicaSet.String(),
	}
	err = k.client.AppsV1().ReplicaSets(namespace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, listReplicaSetOptions)
	if err != nil {
		if !errors.IsNotFound(err) {
			log.Printf("Error when deleting by label. Error was %s\n", err)
			return err
		}
	}

	labelDaemonSet := labels.SelectorFromSet(labels.Set(map[string]string{label: "True"}))
	listDaemonSetOptions := metav1.ListOptions{
		LabelSelector: labelDaemonSet.String(),
	}
	err = k.client.AppsV1().DaemonSets(namespace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, listDaemonSetOptions)
	if err != nil {
		if !errors.IsNotFound(err) {
			log.Printf("Error when deleting by label. Error was %s\n", err)
			return err
		}
	}

	labelStatefulSet := labels.SelectorFromSet(labels.Set(map[string]string{label: "True"}))
	listStatefulSetOptions := metav1.ListOptions{
		LabelSelector: labelStatefulSet.String(),
	}
	err = k.client.AppsV1().StatefulSets(namespace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, listStatefulSetOptions)

	if err != nil {
		if !errors.IsNotFound(err) {
			log.Printf("Error when deleting by label. Error was %s\n", err)
			return err
		}
	}

	labelPod := labels.SelectorFromSet(labels.Set(map[string]string{label: "True"}))
	listPodOptions := metav1.ListOptions{
		LabelSelector: labelPod.String(),
	}
	err = k.client.CoreV1().Pods(namespace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, listPodOptions)
	if err != nil {
		if !errors.IsNotFound(err) {
			log.Printf("Error when deleting by label. Error was %s\n", err)
			return err
		}
	}

	return nil
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
