package serializer

import (
	"log"
	"encoding/json"

	"github.com/greenopsinc/util/auditlog"
	"github.com/greenopsinc/util/clientrequest"
	"github.com/greenopsinc/util/cluster"
	"github.com/greenopsinc/util/event"
	"github.com/greenopsinc/util/git"
	"github.com/greenopsinc/util/serializerutil"
	"github.com/greenopsinc/util/team"
)

func Serialize(object interface{}) string {
	var err error
	var bytes []byte
	switch object.(type) {
	case []auditlog.Log:
		bytes, err = json.Marshal(auditlog.MarshalList(object.([]auditlog.Log)))
	case auditlog.Log:
		bytes, err = auditlog.Marshal(object.(auditlog.Log)).MarshalJSON()
	case git.GitRepoSchema:
		bytes, err = git.MarshalGitRepoSchema(object.(git.GitRepoSchema)).MarshalJSON()
	case team.TeamSchema:
		bytes, err = team.MarshalTeamSchema(object.(team.TeamSchema)).MarshalJSON()
	case git.GitCred:
		bytes, err = git.MarshalGitCred(object.(git.GitCred)).MarshalJSON()
	case event.Event:
		bytes, err = json.Marshal(event.MarshalEvent(object.(event.Event)))
	case clientrequest.NotificationRequestEvent:
		bytes, err = json.Marshal(clientrequest.MarshalNotificationEvent(object.(clientrequest.NotificationRequestEvent)))
	case clientrequest.ClientRequestPacket:
		bytes, err = json.Marshal(clientrequest.MarshalRequestPacket(object.(clientrequest.ClientRequestPacket)))
	case []clientrequest.ClientRequestEvent:
		bytes, err = json.Marshal(clientrequest.MarshalRequestEventList(object.([]clientrequest.ClientRequestEvent)))
	case clientrequest.ClientRequestEvent:
		bytes, err = json.Marshal(clientrequest.MarshalRequestEvent(object.(clientrequest.ClientRequestEvent)))
	case string:
		return object.(string)
	default:
		bytes, err = json.Marshal(object)
	}
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func Deserialize(payload string, deserializeType string) interface{} {
	var err error
	var returnVal interface{}
	if deserializeType == serializerutil.LogType {
		returnVal = auditlog.UnmarshallString(payload)
	} else if deserializeType == serializerutil.PipelineTriggerEventType {
		var pipelineTriggerEvent event.PipelineTriggerEvent
		err = json.Unmarshal([]byte(payload), &pipelineTriggerEvent)
		returnVal = pipelineTriggerEvent
	} else if deserializeType == serializerutil.TeamSchemaType {
		log.Printf(payload)
		returnVal = team.UnmarshallTeamSchemaString(payload)
	} else if deserializeType == serializerutil.NotificationType {
		var notification clientrequest.Notification
		err = json.Unmarshal([]byte(payload), &notification)
		returnVal = notification
	} else if deserializeType == serializerutil.ClusterSchemaType {
		var clusterSchema cluster.ClusterSchema
		err = json.Unmarshal([]byte(payload), &clusterSchema)
		returnVal = clusterSchema
	} else if deserializeType == serializerutil.PipelineInfoType {
		var pipelineInfo auditlog.PipelineInfo
		err = json.Unmarshal([]byte(payload), &pipelineInfo)
		returnVal = pipelineInfo
	} else if deserializeType == serializerutil.StringListType {
		var stringList []string
		err = json.Unmarshal([]byte(payload), &stringList)
		returnVal = stringList
	} else if deserializeType == serializerutil.ClientPacketType {
		returnVal = clientrequest.UnmarshalRequestPacket(payload)
	} else if deserializeType == serializerutil.ClientRequestType {
		returnVal = clientrequest.UnmarshalRequestEvent(payload)
	} else if deserializeType == serializerutil.ClientRequestListType {
		returnVal = clientrequest.UnmarshalRequestEventList(payload)
	} else {
		panic("deserialization types did not match")
	}
	if err != nil {
		panic(err)
	}
	return returnVal
}
