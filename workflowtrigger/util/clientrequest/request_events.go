package clientrequest

import (
	"encoding/json"

	"greenops.io/workflowtrigger/util/serializerutil"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ClientRequestPacket struct {
	RetryCount    int                `json:"retryCount"`
	Namespace     string             `json:"namespace"`
	ClientRequest ClientRequestEvent `json:"clientRequest"`
}

func MarshalRequestPacket(packet ClientRequestPacket) map[string]interface{} {
	bytes, err := json.Marshal(packet)
	if err != nil {
		panic(err)
	}
	var mapObj map[string]interface{}
	err = json.Unmarshal(bytes, &mapObj)
	if err != nil {
		panic(err)
	}
	mapObj["clientRequest"] = MarshalRequestEvent(packet.ClientRequest)
	return mapObj
}

type ClientRequestEvent interface {
	GetRequestEvent() string
}

type NotificationRequestEvent interface {
	GetNotificationEvent() string
	GetRequestId() string
}

type ClientRequestEventMetadata struct {
	OrgName      string `json:"orgName"`
	TeamName     string `json:"teamName"`
	PipelineName string `json:"pipelineName"`
	PipelineUvn  string `json:"pipelineUvn"`
	StepName     string `json:"stepName"`
}

// ClientDeployRequest -----
type ClientDeployRequest struct {
	ClientRequestEventMetadata
	ResponseEventType string `json:"responseEventType"`
	DeployType        string `json:"deployType"`
	RevisionHash      string `json:"revisionHash"`
	Payload           string `json:"payload"`
}

func (r ClientDeployRequest) GetRequestEvent() string {
	return serializerutil.DeployRequestEventType
}

func MarshalRequestEvent(event ClientRequestEvent) map[string]interface{} {
	switch event.(type) {
	case ClientDeployRequest:
		mapObj := serializerutil.GetMapFromStruct(event)
		mapObj["type"] = serializerutil.DeployRequestEventType
		return mapObj
	default:
		panic("Matching notification event type not found")
	}
}

//*****
//*****
//Notifications Events: Need to have a request ID
//*****
//*****

type ClientNotificationEventMetadata struct {
	ClientRequestEventMetadata
	RequestId string `json:"requestId"`
}

// ClientMarkNoDeployRequest -----
type ClientMarkNoDeployRequest struct {
	ClientNotificationEventMetadata
	ClusterName string `json:"clusterName"`
	Namespace   string `json:"namespace"`
	Apply       bool   `json:"apply"`
}

func (r ClientMarkNoDeployRequest) GetNotificationEvent() string {
	return serializerutil.NoDeployNotificationEventType
}

func (r ClientMarkNoDeployRequest) GetRequestId() string {
	return r.RequestId
}

// ClientAggregateRequest -----
type ClientAggregateRequest struct {
	ClientNotificationEventMetadata
	ClusterName string `json:"clusterName"`
	Namespace   string `json:"namespace"`
}

func (r ClientAggregateRequest) GetNotificationEvent() string {
	return serializerutil.AggregateNotificationEventType
}

func (r ClientAggregateRequest) GetRequestId() string {
	return r.RequestId
}

// ClientLabelRequest -----
type ClientLabelRequest struct {
	ClientNotificationEventMetadata
	ResourceList []GvkResourceInfo
}

type GvkGroupRequest struct {
	ResourceList []GvkResourceInfo
}

type GvkResourceInfo struct {
	schema.GroupVersionKind
	ResourceName      string
	ResourceNamespace string
}

func (r ClientLabelRequest) GetNotificationEvent() string {
	return serializerutil.LabelNotificationEventType
}

func (r ClientLabelRequest) GetRequestId() string {
	return r.RequestId
}

func MarshalNotificationEvent(event NotificationRequestEvent) map[string]interface{} {
	switch event.(type) {
	case ClientMarkNoDeployRequest:
		mapObj := serializerutil.GetMapFromStruct(event)
		mapObj["type"] = serializerutil.NoDeployNotificationEventType
		return mapObj
	case ClientAggregateRequest:
		mapObj := serializerutil.GetMapFromStruct(event)
		mapObj["type"] = serializerutil.AggregateNotificationEventType
		return mapObj
	case ClientLabelRequest:
		mapObj := serializerutil.GetMapFromStruct(event)
		mapObj["type"] = serializerutil.LabelNotificationEventType
		return mapObj
	default:
		panic("Matching notification event type not found")
	}
}
