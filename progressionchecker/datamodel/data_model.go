package datamodel

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
}
type WatchKey struct {
	WatchKeyMetaData
	Name string
	//You can't have two Argo apps with the same name in the same namespace, this makes sure there are no collisions
	Namespace                string
	GeneratedCompletionEvent bool
}
