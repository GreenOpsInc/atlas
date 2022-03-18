package clientwrapper

import (
	"encoding/json"

	"github.com/greenopsinc/util/clientrequest"
	"github.com/greenopsinc/util/db"
	"github.com/greenopsinc/workfloworchestrator/ingest/dbkey"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	LatestRevision                = "LATEST_REVISION"
	DeployArgoRequest             = "DeployArgoRequest"
	DeployKubernetesRequest       = "DeployKubernetesRequest"
	DeployTestRequest             = "DeployTestRequest"
	DeleteArgoRequest             = "DeleteArgoRequest"
	DeleteKubernetesRequest       = "DeleteKubernetesRequest"
	DeleteTestRequest             = "DeleteTestRequest"
	ResponseEventApplicationInfra = "ApplicationInfraCompletionEvent"
)

type ClientRequestQueue interface {
	Deploy(clusterName string, orgName string, teamName string, pipelineName string, uvn string, stepName string, stepNamespace string, responseEventType clientrequest.ResponseEventType, objectType string, revisionHash string, payload interface{}) error
	DeployAndWatch(clusterName string, orgName string, teamName string, pipelineName string, uvn string, stepName string, stepNamespace string, deployType string, revisionHash string, payload interface{}, watchType string, testNumber int) error
	SelectiveSyncArgoApplication(clusterName string, orgName string, teamName string, pipelineName string, uvn string, stepName string, stepNamespace string, revisionHash string, resourcesGvkRequest *clientrequest.ResourcesGVKRequest, appName string)
	DeployArgoAppByName(clusterName string, orgName string, pipelineName string, stepName string, appName string)
	DeployArgoAppByNameAndWatch(clusterName string, orgName string, teamName string, pipelineName string, uvn string, stepName string, stepNamespace string, appName string, watchType string)
	RollbackAndWatch(clusterName string, orgName string, teamName string, pipelineName string, uvn string, stepName string, stepNamespace string, appName string, revisionHash string, watchType string)
	DeleteByConfig(clusterName string, orgName string, teamName string, pipelineName string, uvn string, stepName string, stepNamespace string, deleteType string, configPayload string)
	DeleteByGVK(clusterName string, orgName string, teamName string, pipelineName string, uvn string, stepName string, stepNamespace string, deleteType string, resourceName string, resourceNamespace string, group string, version string, kind string)
}

type clientRequestQueue struct {
	dbClient db.DbClient
}

func NewClientRequestQueue(dbClient db.DbClient) ClientRequestQueue {
	return &clientRequestQueue{dbClient: dbClient}
}

func (c *clientRequestQueue) Deploy(clusterName string, orgName string, teamName string, pipelineName string, uvn string, stepName string, stepNamespace string, responseEventType clientrequest.ResponseEventType, objectType string, revisionHash string, payload interface{}) error {
	var body string
	if objectType == DeployTestRequest {
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = string(b)
	} else {
		body = payload.(string)
	}

	deployRequest := clientrequest.ClientDeployRequest{
		ClientRequestEventMetadata: clientrequest.ClientRequestEventMetadata{
			OrgName:      orgName,
			TeamName:     teamName,
			PipelineName: pipelineName,
			PipelineUvn:  uvn,
			StepName:     stepName,
			FinalTry:     false,
		},
		ResponseEventType: responseEventType,
		DeployType:        objectType,
		RevisionHash:      revisionHash,
		Payload:           body,
	}

	dbKey := dbkey.MakeClientRequestQueueKey(orgName, clusterName)
	c.dbClient.InsertValueInTransactionlessList(dbKey, clientrequest.ClientRequestPacket{
		RetryCount:    0,
		Namespace:     stepNamespace,
		ClientRequest: &deployRequest,
	})
	return nil
}

func (c *clientRequestQueue) DeployAndWatch(clusterName string, orgName string, teamName string, pipelineName string, uvn string, stepName string, stepNamespace string, deployType string, revisionHash string, payload interface{}, watchType string, testNumber int) error {
	var body string
	if deployType == DeployTestRequest {
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = string(b)
	} else {
		body = payload.(string)
	}

	deployRequest := clientrequest.ClientDeployAndWatchRequest{
		ClientRequestEventMetadata: clientrequest.ClientRequestEventMetadata{
			OrgName:      orgName,
			TeamName:     teamName,
			PipelineName: pipelineName,
			PipelineUvn:  uvn,
			StepName:     stepName,
			FinalTry:     false,
		},
		TestNumber:   testNumber,
		WatchType:    watchType,
		DeployType:   deployType,
		RevisionHash: revisionHash,
		Payload:      body,
	}

	dbKey := dbkey.MakeClientRequestQueueKey(orgName, clusterName)
	c.dbClient.InsertValueInTransactionlessList(dbKey, clientrequest.ClientRequestPacket{
		RetryCount:    0,
		Namespace:     stepNamespace,
		ClientRequest: &deployRequest,
	})
	return nil
}

func (c *clientRequestQueue) SelectiveSyncArgoApplication(clusterName string, orgName string, teamName string, pipelineName string, uvn string, stepName string, stepNamespace string, revisionHash string, resourcesGvkRequest *clientrequest.ResourcesGVKRequest, appName string) {
	var resList []clientrequest.GvkResourceInfo
	for _, el := range resourcesGvkRequest.ResourceGVKList {
		resList = append(resList, clientrequest.GvkResourceInfo{
			GroupVersionKind: schema.GroupVersionKind{
				Group:   el.Group,
				Version: el.Version,
				Kind:    el.Kind,
			},
			ResourceName:      el.ResourceName,
			ResourceNamespace: el.ResourceNamespace,
		})
	}

	deployRequest := clientrequest.ClientSelectiveSyncRequest{
		ClientRequestEventMetadata: clientrequest.ClientRequestEventMetadata{
			OrgName:      orgName,
			TeamName:     teamName,
			PipelineName: pipelineName,
			PipelineUvn:  uvn,
			StepName:     stepName,
			FinalTry:     false,
		},
		AppName:         appName,
		RevisionHash:    revisionHash,
		GvkResourceList: clientrequest.GvkGroupRequest{ResourceList: resList},
	}

	dbKey := dbkey.MakeClientRequestQueueKey(orgName, clusterName)
	c.dbClient.InsertValueInTransactionlessList(dbKey, clientrequest.ClientRequestPacket{
		RetryCount:    0,
		Namespace:     stepNamespace,
		ClientRequest: &deployRequest,
	})
}

func (c *clientRequestQueue) DeployArgoAppByName(clusterName string, orgName string, pipelineName string, stepName string, appName string) {
	deployRequest := clientrequest.ClientDeployNamedArgoApplicationRequest{
		ClientRequestEventMetadata: clientrequest.ClientRequestEventMetadata{
			OrgName:      orgName,
			PipelineName: pipelineName,
			StepName:     stepName,
			FinalTry:     false,
		},
		DeployType: DeployArgoRequest,
		AppName:    appName,
	}

	dbKey := dbkey.MakeClientRequestQueueKey(orgName, clusterName)
	c.dbClient.InsertValueInTransactionlessList(dbKey, clientrequest.ClientRequestPacket{
		RetryCount:    0,
		Namespace:     "",
		ClientRequest: &deployRequest,
	})
}

func (c *clientRequestQueue) DeployArgoAppByNameAndWatch(clusterName string, orgName string, teamName string, pipelineName string, uvn string, stepName string, stepNamespace string, appName string, watchType string) {
	deployRequest := clientrequest.ClientDeployNamedArgoAppAndWatchRequest{
		ClientRequestEventMetadata: clientrequest.ClientRequestEventMetadata{
			OrgName:      orgName,
			PipelineName: pipelineName,
			PipelineUvn:  uvn,
			StepName:     stepName,
			TeamName:     teamName,
			FinalTry:     false,
		},
		DeployType: DeployArgoRequest,
		AppName:    appName,
		WatchType:  watchType,
	}

	dbKey := dbkey.MakeClientRequestQueueKey(orgName, clusterName)
	c.dbClient.InsertValueInTransactionlessList(dbKey, clientrequest.ClientRequestPacket{
		RetryCount:    0,
		Namespace:     stepNamespace,
		ClientRequest: &deployRequest,
	})
}

func (c *clientRequestQueue) RollbackAndWatch(clusterName string, orgName string, teamName string, pipelineName string, uvn string, stepName string, stepNamespace string, appName string, revisionHash string, watchType string) {
	deployRequest := clientrequest.ClientRollbackAndWatchRequest{
		ClientRequestEventMetadata: clientrequest.ClientRequestEventMetadata{
			OrgName:      orgName,
			PipelineName: pipelineName,
			PipelineUvn:  uvn,
			StepName:     stepName,
			TeamName:     teamName,
			FinalTry:     false,
		},
		AppName:      appName,
		WatchType:    watchType,
		RevisionHash: revisionHash,
	}

	dbKey := dbkey.MakeClientRequestQueueKey(orgName, clusterName)
	c.dbClient.InsertValueInTransactionlessList(dbKey, clientrequest.ClientRequestPacket{
		RetryCount:    0,
		Namespace:     stepNamespace,
		ClientRequest: &deployRequest,
	})
}

func (c *clientRequestQueue) DeleteByConfig(clusterName string, orgName string, teamName string, pipelineName string, uvn string, stepName string, stepNamespace string, deleteType string, configPayload string) {
	deployRequest := clientrequest.ClientDeleteByConfigRequest{
		ClientRequestEventMetadata: clientrequest.ClientRequestEventMetadata{
			OrgName:      orgName,
			PipelineName: pipelineName,
			PipelineUvn:  uvn,
			StepName:     stepName,
			TeamName:     teamName,
			FinalTry:     false,
		},
		DeleteType:    deleteType,
		ConfigPayload: configPayload,
	}

	dbKey := dbkey.MakeClientRequestQueueKey(orgName, clusterName)
	c.dbClient.InsertValueInTransactionlessList(dbKey, clientrequest.ClientRequestPacket{
		RetryCount:    0,
		Namespace:     stepNamespace,
		ClientRequest: &deployRequest,
	})
}

func (c *clientRequestQueue) DeleteByGVK(clusterName string, orgName string, teamName string, pipelineName string, uvn string, stepName string, stepNamespace string, deleteType string, resourceName string, resourceNamespace string, group string, version string, kind string) {
	deployRequest := clientrequest.ClientDeleteByGVKRequest{
		ClientRequestEventMetadata: clientrequest.ClientRequestEventMetadata{
			OrgName:      orgName,
			PipelineName: pipelineName,
			PipelineUvn:  uvn,
			StepName:     stepName,
			TeamName:     teamName,
			FinalTry:     false,
		},
		DeleteType:        deleteType,
		ResourceName:      resourceName,
		ResourceNamespace: resourceNamespace,
		Group:             group,
		Version:           version,
		Kind:              kind,
	}

	dbKey := dbkey.MakeClientRequestQueueKey(orgName, clusterName)
	c.dbClient.InsertValueInTransactionlessList(dbKey, clientrequest.ClientRequestPacket{
		RetryCount:    0,
		Namespace:     stepNamespace,
		ClientRequest: &deployRequest,
	})
}
