package serializer

import (
	"encoding/json"
	"greenops.io/workflowtrigger/pipelinestatus"
	"greenops.io/workflowtrigger/util/auditlog"
	"greenops.io/workflowtrigger/util/cluster"
	"greenops.io/workflowtrigger/util/event"
	"greenops.io/workflowtrigger/util/git"
	"greenops.io/workflowtrigger/util/serializerutil"
	"greenops.io/workflowtrigger/util/team"
)

func Serialize(object interface{}) string {
	var err error
	var bytes []byte
	switch object.(type) {
	case pipelinestatus.PipelineStatus:
		bytes, err = json.Marshal(pipelinestatus.MarshallPipelineStatus(object.(pipelinestatus.PipelineStatus)))
	case []auditlog.Log:
		bytes, err = json.Marshal(auditlog.MarshalList(object.([]auditlog.Log)))
	case auditlog.Log:
		bytes, err = json.Marshal(auditlog.Marshal(object.(auditlog.Log)))
	case git.GitRepoSchema:
		bytes, err = json.Marshal(git.MarshalGitRepoSchema(object.(git.GitRepoSchema)))
	case team.TeamSchema:
		bytes, err = json.Marshal(team.MarshalTeamSchema(object.(team.TeamSchema)))
	case git.GitCred:
		bytes, err = json.Marshal(git.MarshalGitCred(object.(git.GitCred)))
	case event.Event:
		bytes, err = json.Marshal(event.MarshalEvent(object.(event.Event)))
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
		returnVal = team.UnmarshallTeamSchemaString(payload)
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
	} else {
		panic("deserialization types did not match")
	}
	if err != nil {
		panic(err)
	}
	return returnVal
}
