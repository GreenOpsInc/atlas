package argo

import (
	"context"
	"fmt"
	"github.com/argoproj/argo-cd/pkg/apiclient/cluster"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
)

type ArgoClusterApi interface {
	CreateCluster(name string, server string, authConfig v1alpha1.ClusterConfig)
	DeleteCluster(name string, server string)
}

func (a *ArgoApiImpl) CreateCluster(name string, server string, authConfig v1alpha1.ClusterConfig) {
	closer, client, err := a.configuredClient.NewClusterClient()
	if err != nil {
		panic(fmt.Sprintf("cluster client could not be made for Argo: %s", err))
	}
	defer closer.Close()
	_, err = client.Create(context.TODO(), &cluster.ClusterCreateRequest{
		Cluster: &v1alpha1.Cluster{
			Server:     server,
			Name:       name,
			Config:     authConfig,
			Namespaces: nil,
		},
	})
	if err != nil {
		panic(fmt.Sprintf("Creating the cluster failed with error: %s", err))
	}
}

func (a *ArgoApiImpl) DeleteCluster(name string, server string) {
	closer, client, err := a.configuredClient.NewClusterClient()
	if err != nil {
		panic(fmt.Sprintf("cluster client could not be made for Argo: %s", err))
	}
	defer closer.Close()
	_, err = client.Delete(context.TODO(), &cluster.ClusterQuery{
		Server:               server,
		Name:                 name,
	})
	if err != nil {
		panic(fmt.Sprintf("Deleting the cluster failed with error: %s", err))
	}
}
