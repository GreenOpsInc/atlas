package handling

import (
	"github.com/greenopsinc/util/clientrequest"
	"github.com/greenopsinc/util/event"
	"github.com/greenopsinc/util/git"
	"github.com/greenopsinc/util/pipeline/data"
	cw "github.com/greenopsinc/workfloworchestrator/ingest/apiclient/clientwrapper"
	rs "github.com/greenopsinc/workfloworchestrator/ingest/apiclient/reposerver"
)

type DeploymentHandler interface {
	DeleteApplicationInfrastructure(event event.Event, gitRepoSchemaInfo *git.GitRepoSchemaInfo, stepData *data.StepData, gitCommitHash string)
	DeployApplicationInfrastructure(event event.Event, gitRepoSchemaInfo *git.GitRepoSchemaInfo, stepData *data.StepData, gitCommitHash string)
	DeployArgoApplication(event event.Event, gitRepoSchemaInfo *git.GitRepoSchemaInfo, pipelineData *data.PipelineData, stepName string, argoRevisionHash string, gitCommitHash string)
	RollbackArgoApplication(event event.Event, gitRepoSchemaInfo *git.GitRepoSchemaInfo, stepData *data.StepData, argoApplicationName string, argoRevisionHash string)
	TriggerStateRemediation(event event.Event, gitRepoSchemaInfo *git.GitRepoSchemaInfo, stepData *data.StepData, argoApplicationName string, argoRevisionHash string, resourceStatuses []*clientrequest.ResourceGVK)
	RollbackInPipelineExists(event event.Event, pipelineData *data.PipelineData, stepName string)
}

type deploymentHandler struct {
	repoManagerApi       rs.RepoManagerAPI
	clientRequestQueue   cw.ClientRequestQueue
	metadataHandler      MetadataHandler
	deploymentLogHandler DeploymentLogHandler
}

func NewDeploymentHandler(repoManagerApi rs.RepoManagerAPI, clientRequestQueue cw.ClientRequestQueue, metadataHandler MetadataHandler, deploymentLogHandler DeploymentLogHandler) DeploymentHandler {
	return &deploymentHandler{
		repoManagerApi:       repoManagerApi,
		clientRequestQueue:   clientRequestQueue,
		metadataHandler:      metadataHandler,
		deploymentLogHandler: deploymentLogHandler,
	}
}

func (d *deploymentHandler) DeleteApplicationInfrastructure(event event.Event, gitRepoSchemaInfo *git.GitRepoSchemaInfo, stepData *data.StepData, gitCommitHash string) {
	panic("implement me")
}

func (d *deploymentHandler) DeployApplicationInfrastructure(event event.Event, gitRepoSchemaInfo *git.GitRepoSchemaInfo, stepData *data.StepData, gitCommitHash string) {
	panic("implement me")
}

func (d *deploymentHandler) DeployArgoApplication(event event.Event, gitRepoSchemaInfo *git.GitRepoSchemaInfo, pipelineData *data.PipelineData, stepName string, argoRevisionHash string, gitCommitHash string) {
	panic("implement me")
}

func (d *deploymentHandler) RollbackArgoApplication(event event.Event, gitRepoSchemaInfo *git.GitRepoSchemaInfo, stepData *data.StepData, argoApplicationName string, argoRevisionHash string) {
	panic("implement me")
}

func (d *deploymentHandler) TriggerStateRemediation(event event.Event, gitRepoSchemaInfo *git.GitRepoSchemaInfo, stepData *data.StepData, argoApplicationName string, argoRevisionHash string, resourceStatuses []*clientrequest.ResourceGVK) {
	panic("implement me")
}

func (d *deploymentHandler) RollbackInPipelineExists(event event.Event, pipelineData *data.PipelineData, stepName string) {
	panic("implement me")
}
