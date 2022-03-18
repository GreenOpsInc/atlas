package clientrequest

import (
	"encoding/json"

	"github.com/greenopsinc/util/serializerutil"
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

func UnmarshalRequestPacket(str string) ClientRequestPacket {
	var m map[string]interface{}
	err := json.Unmarshal([]byte(str), &m)
	if err != nil {
		panic(err)
	}

	var packet ClientRequestPacket
	packet.RetryCount = int(m["retryCount"].(float64))
	packet.Namespace = m["namespace"].(string)
	packet.ClientRequest = UnmarshalMapRequestEvent(m["clientRequest"].(map[string]interface{}))
	return packet
}

type ClientRequestEvent interface {
	GetEvent() string
	GetClientMetadata() ClientRequestEventMetadata
	GetPipelineUvn() string
	IsFinalTry() bool
	SetFinalTry(finalTry bool)
}
type ClientRequestEventMetadata struct {
	OrgName      string `json:"orgName"`
	TeamName     string `json:"teamName"`
	PipelineName string `json:"pipelineName"`
	PipelineUvn  string `json:"pipelineUvn"`
	StepName     string `json:"stepName"`
	FinalTry     bool   `json:"finalTry"`
}

// ClientDeployRequest -----
type ClientDeployRequest struct {
	ClientRequestEventMetadata
	ResponseEventType ResponseEventType `json:"responseEventType"`
	DeployType        string            `json:"deployType"`
	RevisionHash      string            `json:"revisionHash"`
	Payload           string            `json:"payload"`
}

func (r *ClientDeployRequest) GetEvent() string {
	return serializerutil.ClientDeployRequestType
}

func (r *ClientDeployRequest) GetClientMetadata() ClientRequestEventMetadata {
	return r.ClientRequestEventMetadata
}

func (r *ClientDeployRequest) GetPipelineUvn() string {
	return r.PipelineUvn
}

func (r *ClientDeployRequest) IsFinalTry() bool {
	return r.FinalTry
}
func (r *ClientDeployRequest) SetFinalTry(finalTry bool) {
	r.FinalTry = finalTry
}

// ClientDeleteByConfigRequest -----
type ClientDeleteByConfigRequest struct {
	ClientRequestEventMetadata
	DeleteType    string `json:"deleteType"`
	ConfigPayload string `json:"configPayload"`
}

func (r *ClientDeleteByConfigRequest) GetEvent() string {
	return serializerutil.ClientDeleteByConfigRequestType
}

func (r *ClientDeleteByConfigRequest) GetClientMetadata() ClientRequestEventMetadata {
	return r.ClientRequestEventMetadata
}

func (r *ClientDeleteByConfigRequest) GetPipelineUvn() string {
	return r.PipelineUvn
}

func (r *ClientDeleteByConfigRequest) IsFinalTry() bool {
	return r.FinalTry
}
func (r *ClientDeleteByConfigRequest) SetFinalTry(finalTry bool) {
	r.FinalTry = finalTry
}

// ClientDeleteByGVKRequest -----
type ClientDeleteByGVKRequest struct {
	ClientRequestEventMetadata
	DeleteType        string `json:"deleteType"`
	ResourceName      string `json:"resourceName"`
	ResourceNamespace string `json:"resourceNamespace"`
	Group             string `json:"group"`
	Version           string `json:"version"`
	Kind              string `json:"kind"`
}

func (r *ClientDeleteByGVKRequest) GetEvent() string {
	return serializerutil.ClientDeleteByGvkRequestType
}

func (r *ClientDeleteByGVKRequest) GetClientMetadata() ClientRequestEventMetadata {
	return r.ClientRequestEventMetadata
}

func (r *ClientDeleteByGVKRequest) GetPipelineUvn() string {
	return r.PipelineUvn
}

func (r *ClientDeleteByGVKRequest) IsFinalTry() bool {
	return r.FinalTry
}
func (r *ClientDeleteByGVKRequest) SetFinalTry(finalTry bool) {
	r.FinalTry = finalTry
}

// ClientDeployAndWatchRequest -----
type ClientDeployAndWatchRequest struct {
	ClientRequestEventMetadata
	DeployType   string `json:"deployType"`
	RevisionHash string `json:"revisionHash"`
	Payload      string `json:"payload"`
	WatchType    string `json:"watchType"`
	TestNumber   int    `json:"testNumber"`
}

func (r *ClientDeployAndWatchRequest) GetEvent() string {
	return serializerutil.ClientDeployAndWatchRequestType
}

func (r *ClientDeployAndWatchRequest) GetClientMetadata() ClientRequestEventMetadata {
	return r.ClientRequestEventMetadata
}

func (r *ClientDeployAndWatchRequest) GetPipelineUvn() string {
	return r.PipelineUvn
}

func (r *ClientDeployAndWatchRequest) IsFinalTry() bool {
	return r.FinalTry
}
func (r *ClientDeployAndWatchRequest) SetFinalTry(finalTry bool) {
	r.FinalTry = finalTry
}

// ClientRollbackAndWatchRequest -----
type ClientRollbackAndWatchRequest struct {
	ClientRequestEventMetadata
	AppName      string `json:"appName"`
	RevisionHash string `json:"revisionHash"`
	WatchType    string `json:"watchType"`
}

func (r *ClientRollbackAndWatchRequest) GetEvent() string {
	return serializerutil.ClientRollbackAndWatchRequestType
}

func (r *ClientRollbackAndWatchRequest) GetClientMetadata() ClientRequestEventMetadata {
	return r.ClientRequestEventMetadata
}

func (r *ClientRollbackAndWatchRequest) GetPipelineUvn() string {
	return r.PipelineUvn
}

func (r *ClientRollbackAndWatchRequest) IsFinalTry() bool {
	return r.FinalTry
}
func (r *ClientRollbackAndWatchRequest) SetFinalTry(finalTry bool) {
	r.FinalTry = finalTry
}

// ClientDeployNamedArgoApplicationRequest -----

type ClientDeployNamedArgoApplicationRequest struct {
	ClientRequestEventMetadata
	DeployType string `json:"deployType"`
	AppName    string `json:"appName"`
}

func (r *ClientDeployNamedArgoApplicationRequest) GetEvent() string {
	return serializerutil.ClientRollbackAndWatchRequestType
}

func (r *ClientDeployNamedArgoApplicationRequest) GetClientMetadata() ClientRequestEventMetadata {
	return r.ClientRequestEventMetadata
}

func (r *ClientDeployNamedArgoApplicationRequest) GetPipelineUvn() string {
	return r.PipelineUvn
}

func (r *ClientDeployNamedArgoApplicationRequest) IsFinalTry() bool {
	return r.FinalTry
}

func (r *ClientDeployNamedArgoApplicationRequest) SetFinalTry(finalTry bool) {
	r.FinalTry = finalTry
}

// ClientDeployNamedArgoAppAndWatchRequest -----

type ClientDeployNamedArgoAppAndWatchRequest struct {
	ClientRequestEventMetadata
	DeployType string `json:"deployType"`
	AppName    string `json:"appName"`
	WatchType  string `json:"watchType"`
}

func (r *ClientDeployNamedArgoAppAndWatchRequest) GetEvent() string {
	return serializerutil.ClientRollbackAndWatchRequestType
}

func (r *ClientDeployNamedArgoAppAndWatchRequest) GetClientMetadata() ClientRequestEventMetadata {
	return r.ClientRequestEventMetadata
}

func (r *ClientDeployNamedArgoAppAndWatchRequest) GetPipelineUvn() string {
	return r.PipelineUvn
}

func (r *ClientDeployNamedArgoAppAndWatchRequest) IsFinalTry() bool {
	return r.FinalTry
}

func (r *ClientDeployNamedArgoAppAndWatchRequest) SetFinalTry(finalTry bool) {
	r.FinalTry = finalTry
}

// ClientSelectiveSyncRequest -----
type ClientSelectiveSyncRequest struct {
	ClientRequestEventMetadata
	AppName         string          `json:"appName"`
	RevisionHash    string          `json:"revisionHash"`
	GvkResourceList GvkGroupRequest `json:"resourcesGvkRequest"`
}

func (r *ClientSelectiveSyncRequest) GetEvent() string {
	return serializerutil.ClientSelectiveSyncRequestType
}

func (r *ClientSelectiveSyncRequest) GetClientMetadata() ClientRequestEventMetadata {
	return r.ClientRequestEventMetadata
}

func (r *ClientSelectiveSyncRequest) GetPipelineUvn() string {
	return r.PipelineUvn
}

func (r *ClientSelectiveSyncRequest) IsFinalTry() bool {
	return r.FinalTry
}
func (r *ClientSelectiveSyncRequest) SetFinalTry(finalTry bool) {
	r.FinalTry = finalTry
}

func MarshalRequestEvent(event ClientRequestEvent) map[string]interface{} {
	mapObj := serializerutil.GetMapFromStruct(event)
	mapObj["type"] = event.GetEvent()
	return mapObj
}

func UnmarshalRequestEventList(str string) []ClientRequestEvent {
	var jsonArray []map[string]interface{}
	json.Unmarshal([]byte(str), &jsonArray)
	var arr []ClientRequestEvent
	for _, val := range jsonArray {
		arr = append(arr, UnmarshalMapRequestEvent(val))
	}
	return arr
}

func UnmarshalMapRequestEvent(m map[string]interface{}) ClientRequestEvent {
	logBytes, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	if m["type"] == serializerutil.ClientDeployRequestType {
		var request ClientDeployRequest
		_ = json.Unmarshal(logBytes, &request)
		return &request
	} else if m["type"] == serializerutil.ClientDeleteByConfigRequestType {
		var request ClientDeleteByConfigRequest
		_ = json.Unmarshal(logBytes, &request)
		return &request
	} else if m["type"] == serializerutil.ClientDeleteByGvkRequestType {
		var request ClientDeleteByGVKRequest
		_ = json.Unmarshal(logBytes, &request)
		return &request
	} else if m["type"] == serializerutil.ClientDeployAndWatchRequestType {
		var request ClientDeployAndWatchRequest
		_ = json.Unmarshal(logBytes, &request)
		return &request
	} else if m["type"] == serializerutil.ClientRollbackAndWatchRequestType {
		var request ClientRollbackAndWatchRequest
		_ = json.Unmarshal(logBytes, &request)
		return &request
	} else if m["type"] == serializerutil.ClientSelectiveSyncRequestType {
		var request ClientSelectiveSyncRequest
		_ = json.Unmarshal(logBytes, &request)
		return &request
	} else if m["type"] == serializerutil.ClientMarkNoDeployRequestType {
		var request ClientMarkNoDeployRequest
		_ = json.Unmarshal(logBytes, &request)
		return &request
	} else if m["type"] == serializerutil.ClientLabelRequestType {
		var request ClientLabelRequest
		_ = json.Unmarshal(logBytes, &request)
		return &request
	} else if m["type"] == serializerutil.ClientAggregateRequestType {
		var request ClientAggregateRequest
		_ = json.Unmarshal(logBytes, &request)
		return &request
	} else if m["type"] == serializerutil.ClientDeleteByLabelRequestType {
		var request ClientDeleteByLabelRequest
		_ = json.Unmarshal(logBytes, &request)
		return &request
	} else {
		panic("Request event type did not match any supported type")
	}
}

func UnmarshalRequestEvent(str string) ClientRequestEvent {
	var m map[string]interface{}
	_ = json.Unmarshal([]byte(str), &m)
	return UnmarshalMapRequestEvent(m)
}

func MarshalRequestEventList(eventList []ClientRequestEvent) []map[string]interface{} {
	var mapArr []map[string]interface{}
	for _, val := range eventList {
		mapArr = append(mapArr, MarshalRequestEvent(val))
	}
	return mapArr
}

//*****
//*****
//Notifications Events: Need to have a request ID
//*****
//*****

type NotificationRequestEvent interface {
	GetEvent() string
	GetRequestId() string
	SetRequestId(id string)
}

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

func (r *ClientMarkNoDeployRequest) GetEvent() string {
	return serializerutil.ClientMarkNoDeployRequestType
}

func (r *ClientMarkNoDeployRequest) GetClientMetadata() ClientRequestEventMetadata {
	return r.ClientNotificationEventMetadata.ClientRequestEventMetadata
}

func (r *ClientMarkNoDeployRequest) GetPipelineUvn() string {
	return r.PipelineUvn
}

func (r *ClientMarkNoDeployRequest) IsFinalTry() bool {
	return r.FinalTry
}
func (r *ClientMarkNoDeployRequest) SetFinalTry(finalTry bool) {
	r.FinalTry = finalTry
}

func (r *ClientMarkNoDeployRequest) GetRequestId() string {
	return r.RequestId
}

func (r *ClientMarkNoDeployRequest) SetRequestId(id string) {
	r.RequestId = id
}

// ClientLabelRequest -----
type ClientLabelRequest struct {
	ClientNotificationEventMetadata
	GvkResourceList GvkGroupRequest `json:"resourcesGvkRequest"`
	RequestId       string          `json:"requestId"`
}

func (r *ClientLabelRequest) GetEvent() string {
	return serializerutil.ClientLabelRequestType
}

func (r *ClientLabelRequest) GetClientMetadata() ClientRequestEventMetadata {
	return r.ClientNotificationEventMetadata.ClientRequestEventMetadata
}

func (r *ClientLabelRequest) GetPipelineUvn() string {
	return r.PipelineUvn
}

func (r *ClientLabelRequest) IsFinalTry() bool {
	return r.FinalTry
}
func (r *ClientLabelRequest) SetFinalTry(finalTry bool) {
	r.FinalTry = finalTry
}

func (r *ClientLabelRequest) GetRequestId() string {
	return r.RequestId
}

func (r *ClientLabelRequest) SetRequestId(id string) {
	r.RequestId = id
}

// ClientAggregateRequest -----
type ClientAggregateRequest struct {
	ClientNotificationEventMetadata
	ClusterName string `json:"clusterName"`
	Namespace   string `json:"namespace"`
	RequestId   string `json:"requestId"`
}

func (r *ClientAggregateRequest) GetEvent() string {
	return serializerutil.ClientAggregateRequestType
}

func (r *ClientAggregateRequest) GetClientMetadata() ClientRequestEventMetadata {
	return r.ClientNotificationEventMetadata.ClientRequestEventMetadata
}

func (r *ClientAggregateRequest) GetPipelineUvn() string {
	return r.PipelineUvn
}

func (r *ClientAggregateRequest) IsFinalTry() bool {
	return r.FinalTry
}
func (r *ClientAggregateRequest) SetFinalTry(finalTry bool) {
	r.FinalTry = finalTry
}

func (r *ClientAggregateRequest) GetRequestId() string {
	return r.RequestId
}

func (r *ClientAggregateRequest) SetRequestId(id string) {
	r.RequestId = id
}

// ClientDeleteByLabelRequest -----
type ClientDeleteByLabelRequest struct {
	ClientNotificationEventMetadata
	Namespace string `json:"namespace"`
	RequestId string `json:"requestId"`
}

func (r *ClientDeleteByLabelRequest) GetEvent() string {
	return serializerutil.ClientDeleteByLabelRequestType
}

func (r *ClientDeleteByLabelRequest) GetClientMetadata() ClientRequestEventMetadata {
	return r.ClientNotificationEventMetadata.ClientRequestEventMetadata
}

func (r *ClientDeleteByLabelRequest) GetPipelineUvn() string {
	return r.PipelineUvn
}

func (r *ClientDeleteByLabelRequest) IsFinalTry() bool {
	return r.FinalTry
}
func (r *ClientDeleteByLabelRequest) SetFinalTry(finalTry bool) {
	r.FinalTry = finalTry
}

func (r *ClientDeleteByLabelRequest) GetRequestId() string {
	return r.RequestId
}

func (r *ClientDeleteByLabelRequest) SetRequestId(id string) {
	r.RequestId = id
}

func MarshalNotificationEvent(event NotificationRequestEvent) map[string]interface{} {
	mapObj := serializerutil.GetMapFromStruct(event)
	mapObj["type"] = event.GetEvent()
	return mapObj
}
