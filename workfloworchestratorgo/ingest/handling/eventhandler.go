package handling

type WatchKeyType string

const (
	//Core
	WatchArgoApplicationKey WatchKeyType = "WatchArgoApplicationKey"
	WatchTestKey            WatchKeyType = "WatchTestKey"
	//Plugin
	WatchArgoWorkflowKey WatchKeyType = "ArgoWorkflowTask"
)
