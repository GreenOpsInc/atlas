package datamodel

import "strings"

const (
	Missing    string = "Missing"
	NotHealthy string = "NotHealthy"
	Healthy    string = "Healthy"
)

type ArgoAppMetricInfo struct {
	DestNamespace string `tag:"dest_namespace"`
	DestServer    string `tag:"dest_server"`
	HealthStatus  string `tag:"health_status"`
	Name          string `tag:"name"`
	Namespace     string `tag:"namespace"`
	Operation     string `tag:"operation"`
	Project       string `tag:"project"`
	Repo          string `tag:"repo"`
	SyncStatus    string `tag:"sync_status"`
}

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
	Status                   string
	GeneratedCompletionEvent bool
}

func (key WatchKey) GetKeyFromWatchKey() string {
	return strings.Join([]string{key.TeamName, key.PipelineName, key.StepName, key.Name, key.Namespace}, "-")
}
