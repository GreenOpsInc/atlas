package handling

type WatchKeyType string

const (
	WatchArgoApplicationKey WatchKeyType = "WatchArgoApplicationKey"
	WatchTestKey            WatchKeyType = "WatchTestKey"
	WatchArgoWorkflowKey    WatchKeyType = "ArgoWorkflowTask"
)
