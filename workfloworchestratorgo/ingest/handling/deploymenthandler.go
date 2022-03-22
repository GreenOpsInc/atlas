package handling

import (
	"errors"
	"log"

	"github.com/greenopsinc/util/clientrequest"
	"github.com/greenopsinc/util/event"
	"github.com/greenopsinc/util/git"
	"github.com/greenopsinc/util/pipeline/data"
	cw "github.com/greenopsinc/workfloworchestrator/ingest/apiclient/clientwrapper"
	rs "github.com/greenopsinc/workfloworchestrator/ingest/apiclient/reposerver"
	"gopkg.in/yaml.v2"
)

type DeploymentHandler interface {
	DeleteApplicationInfrastructure(e event.Event, gitRepoSchemaInfo *git.GitRepoSchemaInfo, stepData *data.StepData, gitCommitHash string) error
	DeployApplicationInfrastructure(e event.Event, gitRepoSchemaInfo *git.GitRepoSchemaInfo, stepData *data.StepData, gitCommitHash string) error
	DeployArgoApplication(e event.Event, gitRepoSchemaInfo *git.GitRepoSchemaInfo, pipelineData *data.PipelineData, stepName string, argoRevisionHash string, gitCommitHash string) error
	RollbackArgoApplication(e event.Event, gitRepoSchemaInfo *git.GitRepoSchemaInfo, stepData *data.StepData, argoApplicationName string, argoRevisionHash string) error
	TriggerStateRemediation(e event.Event, gitRepoSchemaInfo *git.GitRepoSchemaInfo, stepData *data.StepData, argoApplicationName string, argoRevisionHash string, resourceStatuses []*clientrequest.ResourceGVK) error
	RollbackInPipelineExists(e event.Event, pipelineData *data.PipelineData, stepName string) bool
}

type deploymentHandler struct {
	repoManagerAPI       rs.RepoManagerAPI
	clientRequestQueue   cw.ClientRequestQueue
	metadataHandler      MetadataHandler
	deploymentLogHandler DeploymentLogHandler
}

func NewDeploymentHandler(repoManagerApi rs.RepoManagerAPI, clientRequestQueue cw.ClientRequestQueue, metadataHandler MetadataHandler, deploymentLogHandler DeploymentLogHandler) DeploymentHandler {
	return &deploymentHandler{
		repoManagerAPI:       repoManagerApi,
		clientRequestQueue:   clientRequestQueue,
		metadataHandler:      metadataHandler,
		deploymentLogHandler: deploymentLogHandler,
	}
}

func (d *deploymentHandler) DeleteApplicationInfrastructure(e event.Event, gitRepoSchemaInfo *git.GitRepoSchemaInfo, stepData *data.StepData, gitCommitHash string) error {
	if stepData.OtherDeploymentsPath != "" {
		getFileRequest := &git.GetFileRequest{
			GitRepoSchemaInfo: *gitRepoSchemaInfo,
			Filename:          e.GetOrgName(),
			GitCommitHash:     gitCommitHash,
		}
		otherDeploymentsConfig, err := d.repoManagerAPI.GetFileFromRepo(getFileRequest, e.GetOrgName(), e.GetTeamName())
		if err != nil {
			return err
		}
		log.Println("deleting old application infrastructure...")

		if otherDeploymentsConfig == "" {
			return nil
		}

		stepNS, err := GetStepNamespace(e, d.repoManagerAPI, stepData.ArgoApplicationPath, gitRepoSchemaInfo, gitCommitHash)
		if err != nil {
			return err
		}
		d.clientRequestQueue.DeleteByConfig(
			stepData.ClusterName,
			e.GetOrgName(),
			e.GetTeamName(),
			e.GetPipelineName(),
			e.GetUVN(),
			e.GetStepName(),
			stepNS,
			clientrequest.DeleteKubernetesRequest,
			otherDeploymentsConfig,
		)
	}
	return nil
}

func (d *deploymentHandler) DeployApplicationInfrastructure(e event.Event, gitRepoSchemaInfo *git.GitRepoSchemaInfo, stepData *data.StepData, gitCommitHash string) error {
	if stepData.OtherDeploymentsPath != "" {
		getFileRequest := &git.GetFileRequest{
			GitRepoSchemaInfo: *gitRepoSchemaInfo,
			Filename:          stepData.OtherDeploymentsPath,
			GitCommitHash:     gitCommitHash,
		}
		otherDeploymentsConfig, err := d.repoManagerAPI.GetFileFromRepo(getFileRequest, e.GetOrgName(), e.GetTeamName())
		if err != nil {
			return err
		}
		log.Println("deploying old application infrastructure...")

		if otherDeploymentsConfig == "" {
			return nil
		}

		stepNS, err := GetStepNamespace(e, d.repoManagerAPI, stepData.ArgoApplicationPath, gitRepoSchemaInfo, gitCommitHash)
		if err != nil {
			return err
		}
		if err = d.clientRequestQueue.Deploy(
			stepData.ClusterName,
			e.GetOrgName(),
			e.GetTeamName(),
			e.GetPipelineName(),
			e.GetUVN(),
			stepData.Name,
			stepNS,
			clientrequest.ResponseEventApplicationInfra,
			clientrequest.DeployKubernetesRequest,
			clientrequest.LatestRevision,
			otherDeploymentsConfig,
		); err != nil {
			return err
		}
	}
	return nil
}

func (d *deploymentHandler) DeployArgoApplication(e event.Event, gitRepoSchemaInfo *git.GitRepoSchemaInfo, pipelineData *data.PipelineData, stepName string, argoRevisionHash string, gitCommitHash string) error {
	stepData := pipelineData.GetStep(stepName)
	if stepData.OtherDeploymentsPath != "" {
		getFileRequest := &git.GetFileRequest{
			GitRepoSchemaInfo: *gitRepoSchemaInfo,
			Filename:          stepData.ArgoApplicationPath,
			GitCommitHash:     gitCommitHash,
		}
		argoApplicationConfig, err := d.repoManagerAPI.GetFileFromRepo(getFileRequest, e.GetOrgName(), e.GetTeamName())
		if err != nil {
			return err
		}
		if err = d.metadataHandler.AssertArgoRepoMetadataExists(e, stepData.Name, argoApplicationConfig); err != nil {
			return err
		}
		pipelineLockRevisionHash := clientrequest.LatestRevision
		if pipelineData.ArgoVersionLock {
			pipelineLockRevisionHash = d.metadataHandler.GetPipelineLockRevisionHash(e, pipelineData, stepName)
		}

		stepNS, err := GetStepNamespace(e, d.repoManagerAPI, stepData.ArgoApplicationPath, gitRepoSchemaInfo, gitCommitHash)
		if err != nil {
			return err
		}
		if err = d.clientRequestQueue.DeployAndWatch(
			stepData.ClusterName,
			e.GetOrgName(),
			e.GetTeamName(),
			e.GetPipelineName(),
			e.GetUVN(),
			stepData.Name,
			stepNS,
			clientrequest.DeployArgoRequest,
			pipelineLockRevisionHash,
			argoApplicationConfig,
			string(WatchArgoApplicationKey),
			-1,
		); err != nil {
			return err
		}
		log.Println("deploying and watching Argo application...")
	}
	return nil
}

func (d *deploymentHandler) RollbackArgoApplication(e event.Event, gitRepoSchemaInfo *git.GitRepoSchemaInfo, stepData *data.StepData, argoApplicationName string, argoRevisionHash string) error {
	stepNS, err := GetStepNamespace(e, d.repoManagerAPI, stepData.ArgoApplicationPath, gitRepoSchemaInfo, argoRevisionHash)
	if err != nil {
		return err
	}
	d.clientRequestQueue.RollbackAndWatch(
		stepData.ClusterName,
		e.GetOrgName(),
		e.GetTeamName(),
		e.GetPipelineName(),
		e.GetUVN(),
		stepData.Name,
		stepNS,
		argoApplicationName,
		argoRevisionHash,
		string(WatchArgoApplicationKey),
	)
	return nil
}

func (d *deploymentHandler) TriggerStateRemediation(e event.Event, gitRepoSchemaInfo *git.GitRepoSchemaInfo, stepData *data.StepData, argoApplicationName string, argoRevisionHash string, resourceStatuses []*clientrequest.ResourceGVK) error {
	syncRequestPayload := &clientrequest.ResourcesGVKRequest{
		resourceStatuses,
	}
	stepNS, err := GetStepNamespace(e, d.repoManagerAPI, stepData.ArgoApplicationPath, gitRepoSchemaInfo, argoRevisionHash)
	if err != nil {
		return err
	}
	d.clientRequestQueue.SelectiveSyncArgoApplication(
		stepData.ClusterName,
		e.GetOrgName(),
		e.GetTeamName(),
		e.GetPipelineName(),
		e.GetUVN(),
		stepData.Name,
		stepNS,
		argoRevisionHash,
		syncRequestPayload,
		argoApplicationName,
	)
	return nil
}

func (d *deploymentHandler) RollbackInPipelineExists(e event.Event, pipelineData *data.PipelineData, stepName string) bool {
	matchingSteps := d.metadataHandler.FindAllStepsWithSameArgoRepoSrc(e, pipelineData, stepName)
	for _, step := range matchingSteps {
		latestDeploymentLog := d.deploymentLogHandler.GetLatestDeploymentLog(e, step)
		if latestDeploymentLog == nil && (latestDeploymentLog.GetUniqueVersionInstance() > 0 || latestDeploymentLog.RollbackUniqueVersionNumber != "") {
			return true
		}
	}
	return false
}

func GetStepNamespace(e event.Event, repoManagerApi rs.RepoManagerAPI, argoApplicationPath string, gitRepoSchemaInfo *git.GitRepoSchemaInfo, gitCommitHash string) (string, error) {
	if e.GetStepName() == "" || e.GetStepName() == data.RootStepName {
		return "", errors.New("could not find a namespace associated with the event")
	}
	if argoApplicationPath == "" {
		return data.DefaultNamespace, nil
	}
	getFileRequest := &git.GetFileRequest{
		GitRepoSchemaInfo: *gitRepoSchemaInfo,
		Filename:          argoApplicationPath,
		GitCommitHash:     gitCommitHash,
	}
	argoAppPayload, err := repoManagerApi.GetFileFromRepo(getFileRequest, e.GetOrgName(), e.GetTeamName())
	if err != nil {
		return "", err
	}
	rawPayload := make(map[interface{}]interface{}, 0)
	if err = yaml.Unmarshal([]byte(argoAppPayload), &rawPayload); err != nil {
		return "", err
	}
	dest, ok := rawPayload["spec"].(map[interface{}]interface{})["destination"].(map[interface{}]interface{})
	if !ok {
		return "", errors.New("failed to parse argoAppPayload")
	}

	ns, ok := dest["namespace"].(string)
	if !ok {
		return "", errors.New("could not find a namespace associated with the event")
	}
	return ns, nil
}
