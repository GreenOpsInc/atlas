package datamodel

import (
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/argoproj/gitops-engine/pkg/health"
	"strings"
)

type ArgoAppMetricInfo struct {
	DestNamespace string                  `tag:"dest_namespace"`
	DestServer    string                  `tag:"dest_server"`
	HealthStatus  health.HealthStatusCode `tag:"health_status"`
	Name          string                  `tag:"name"`
	Namespace     string                  `tag:"namespace"`
	Operation     string                  `tag:"operation"`
	Project       string                  `tag:"project"`
	Repo          string                  `tag:"repo"`
	SyncStatus    v1alpha1.SyncStatusCode `tag:"sync_status"`
}

type StatusType string

const (
	HealthStatus StatusType = "HealthStatus"
	SyncStatus   StatusType = "SyncStatus"
)

type WatchKeyType string

const (
	WatchArgoApplicationKey WatchKeyType = "WatchArgoApplicationKey"
	WatchTestKey            WatchKeyType = "WatchTestKey"
)

type WatchKeyMetaData struct {
	Type         WatchKeyType
	OrgName      string
	TeamName     string
	PipelineName string
	StepName     string
	TestNumber   int
}
type WatchKey struct {
	WatchKeyMetaData
	Name string
	//You can't have two Argo apps with the same name in the same namespace, this makes sure there are no collisions
	Namespace                string
	HealthStatus             string
	SyncStatus               string
	GeneratedCompletionEvent bool
	Resources                *[]ResourceStatus
}

func (key WatchKey) GetKeyFromWatchKey() string {
	return strings.Join([]string{key.TeamName, key.PipelineName, key.StepName, key.Name, key.Namespace}, "-")
}
