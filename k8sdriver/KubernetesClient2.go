package main

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type KubernetesClient interface {
	//TODO: Add parameters for Deploy
	Deploy() bool
	//TODO: Add parameters for Delete
	Delete() bool
	//TODO: Update parameters & return type for CheckStatus
	CheckStatus() bool
	//TODO: Update parameters for ExecInPod
	ExecInPod() bool
}

type KubernetesClientDriver struct {
	clientset *kubernetes.Clientset
}

//TODO: ALL functions should have a callee tag on them
func New() KubernetesClient {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	var client KubernetesClient
	client = KubernetesClientDriver{clientset}
	return client
}

func (k KubernetesClientDriver) Deploy() bool {
	panic("implement me")
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
