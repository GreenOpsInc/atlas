package datamodel

type EventInfoType string

const (
	ClientCompletionEvent EventInfoType = "clientcompletion"
	TestCompletionEvent   EventInfoType = "testcompletion"
)

type EventInfo interface {
	GetEventType() EventInfoType
}

type EventInfoMetaData struct {
	Type         EventInfoType `json:"type"`
	OrgName      string        `json:"orgName"`
	TeamName     string        `json:"teamName"`
	PipelineName string        `json:"pipelineName"`
	StepName     string        `json:"stepName"`
}

type ApplicationEventInfo struct {
	EventInfoMetaData
	HealthStatus string `json:"healthStatus"`
	ArgoName     string `json:"argoName"`
	Operation    string `json:"operation"`
	Project      string `json:"project"`
	Repo         string `json:"repo"`
}

type TestEventInfo struct {
	EventInfoMetaData
	Successful bool   `json:"successful"`
	Log        string `json:"log"`
	TestName   string `json:"testName"`
}

// **
//ApplicationEventInfo functionality
// **
func (eventInfo ApplicationEventInfo) GetEventType() EventInfoType {
	return eventInfo.Type
}

func MakeApplicationEvent(key WatchKey, appInfo ArgoAppMetricInfo) EventInfo {
	return ApplicationEventInfo{
		EventInfoMetaData: EventInfoMetaData{
			Type:         ClientCompletionEvent,
			OrgName:      key.OrgName,
			TeamName:     key.TeamName,
			PipelineName: key.PipelineName,
			StepName:     key.StepName,
		},
		HealthStatus: Healthy,
		ArgoName:     appInfo.Name,
		Operation:    appInfo.Operation,
		Project:      appInfo.Project,
		Repo:         appInfo.Repo,
	}
}

// **
//TestEventInfo functionality
// **
func (eventInfo TestEventInfo) GetEventType() EventInfoType {
	return eventInfo.Type
}

func MakeTestEvent(key WatchKey, successful bool, logs string) EventInfo {
	return TestEventInfo{
		EventInfoMetaData: EventInfoMetaData{
			Type:         TestCompletionEvent,
			OrgName:      key.OrgName,
			TeamName:     key.TeamName,
			PipelineName: key.PipelineName,
			StepName:     key.StepName,
		},
		Successful: successful,
		Log:        logs,
		TestName:   key.Name,
	}
}
