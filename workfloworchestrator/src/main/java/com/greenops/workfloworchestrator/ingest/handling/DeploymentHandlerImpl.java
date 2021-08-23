package com.greenops.workfloworchestrator.ingest.handling;

import com.greenops.util.datamodel.event.Event;
import com.greenops.util.datamodel.event.ResourceStatus;
import com.greenops.util.datamodel.request.GetFileRequest;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.PipelineData;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.StepData;
import com.greenops.workfloworchestrator.datamodel.requests.ResourceGvk;
import com.greenops.workfloworchestrator.datamodel.requests.ResourcesGvkRequest;
import com.greenops.workfloworchestrator.datamodel.requests.WatchRequest;
import com.greenops.workfloworchestrator.error.AtlasRetryableError;
import com.greenops.workfloworchestrator.ingest.apiclient.clientwrapper.ClientWrapperApi;
import com.greenops.workfloworchestrator.ingest.apiclient.reposerver.RepoManagerApi;
import com.greenops.workfloworchestrator.ingest.handling.util.deployment.ArgoDeploymentInfo;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.util.List;

import static com.greenops.workfloworchestrator.ingest.handling.EventHandlerImpl.WATCH_ARGO_APPLICATION_KEY;
import static com.greenops.workfloworchestrator.ingest.handling.util.deployment.ArgoDeploymentInfo.NO_OP_ARGO_DEPLOYMENT;

@Slf4j
@Component
public class DeploymentHandlerImpl implements DeploymentHandler {

    private RepoManagerApi repoManagerApi;
    private ClientWrapperApi clientWrapperApi;
    private MetadataHandler metadataHandler;
    private DeploymentLogHandler deploymentLogHandler;

    @Autowired
    DeploymentHandlerImpl(RepoManagerApi repoManagerApi, ClientWrapperApi clientWrapperApi, MetadataHandler metadataHandler, DeploymentLogHandler deploymentLogHandler) {
        this.repoManagerApi = repoManagerApi;
        this.clientWrapperApi = clientWrapperApi;
        this.metadataHandler = metadataHandler;
        this.deploymentLogHandler = deploymentLogHandler;
    }

    @Override
    public void deleteApplicationInfrastructure(Event event, String pipelineRepoUrl, StepData stepData, String gitCommitHash) {
        if (stepData.getOtherDeploymentsPath() != null) {
            var getFileRequest = new GetFileRequest(pipelineRepoUrl, stepData.getOtherDeploymentsPath(), gitCommitHash);
            var otherDeploymentsConfig = repoManagerApi.getFileFromRepo(getFileRequest, event.getOrgName(), event.getTeamName());
            log.info("Deleting old application infrastructure...");
            for (var deploymentConfig : otherDeploymentsConfig.split("---")) {
                if (deploymentConfig.isBlank()) continue;
                clientWrapperApi.delete(stepData.getClusterName(), event.getOrgName(), ClientWrapperApi.DELETE_KUBERNETES_REQUEST, deploymentConfig);
            }
        }
    }

    @Override
    public void deployApplicationInfrastructure(Event event, String pipelineRepoUrl, StepData stepData, String gitCommitHash) {
        if (stepData.getOtherDeploymentsPath() != null) {
            var getFileRequest = new GetFileRequest(pipelineRepoUrl, stepData.getOtherDeploymentsPath(), gitCommitHash);
            var otherDeploymentsConfig = repoManagerApi.getFileFromRepo(getFileRequest, event.getOrgName(), event.getTeamName());
            log.info("Deploying new application infrastructure...");
            for (var deploymentConfig : otherDeploymentsConfig.split("---")) {
                if (deploymentConfig.isBlank()) continue;
                var deployResponse = clientWrapperApi.deploy(stepData.getClusterName(), event.getOrgName(), ClientWrapperApi.DEPLOY_KUBERNETES_REQUEST, ClientWrapperApi.LATEST_REVISION, deploymentConfig);
                if (!deployResponse.getSuccess()) {
                    var message = "Deploying other resources failed.";
                    log.error(message);
                    throw new AtlasRetryableError(message);
                }
            }
        }
    }

    @Override
    public ArgoDeploymentInfo deployArgoApplication(Event event, String pipelineRepoUrl, PipelineData pipelineData, String stepName, String argoRevisionHash, String gitCommitHash) {
        var stepData = pipelineData.getStep(stepName);
        if (stepData.getArgoApplicationPath() != null) {
            var getFileRequest = new GetFileRequest(pipelineRepoUrl, stepData.getArgoApplicationPath(), gitCommitHash);
            var argoApplicationConfig = repoManagerApi.getFileFromRepo(getFileRequest, event.getOrgName(), event.getTeamName());
            metadataHandler.assertArgoRepoMetadataExists(event, stepData.getName(), argoApplicationConfig);
            var pipelineLockRevisionHash = pipelineData.isArgoVersionLock() ? metadataHandler.getPipelineLockRevisionHash(event, pipelineData, stepData.getName()) : null;
            pipelineLockRevisionHash = pipelineLockRevisionHash != null ? pipelineLockRevisionHash : ClientWrapperApi.LATEST_REVISION;
            var deployResponse = clientWrapperApi.deploy(stepData.getClusterName(), event.getOrgName(), ClientWrapperApi.DEPLOY_ARGO_REQUEST, pipelineLockRevisionHash, argoApplicationConfig);
            log.info("Deploying Argo application {}...", deployResponse.getResourceName());
            if (!deployResponse.getSuccess()) {
                var message = "Deploying the Argo application failed.";
                log.error(message);
                throw new AtlasRetryableError(message);
            } else {
                var watchRequest = new WatchRequest(event.getTeamName(), event.getPipelineName(), stepData.getName(), WATCH_ARGO_APPLICATION_KEY, deployResponse.getResourceName(), deployResponse.getApplicationNamespace());
                clientWrapperApi.watchApplication(stepData.getClusterName(), event.getOrgName(), watchRequest);
                log.info("Watching Argo application {}", deployResponse.getResourceName());
                return new ArgoDeploymentInfo(deployResponse.getResourceName(), deployResponse.getRevisionHash());
            }

        } else { //stepData.getArgoApplication() != null
            var deployResponse = clientWrapperApi.deployArgoAppByName(stepData.getClusterName(), event.getOrgName(), stepData.getArgoApplication());
            log.info("Syncing the Argo application {}...", deployResponse.getResourceName());
            if (!deployResponse.getSuccess()) {
                var message = "Syncing the Argo application failed.";
                log.error(message);
                throw new AtlasRetryableError(message);
            } else {
                var watchRequest = new WatchRequest(event.getTeamName(), event.getPipelineName(), stepData.getName(), WATCH_ARGO_APPLICATION_KEY, deployResponse.getResourceName(), deployResponse.getApplicationNamespace());
                clientWrapperApi.watchApplication(stepData.getClusterName(), event.getOrgName(), watchRequest);
                log.info("Watching Argo application {}", deployResponse.getResourceName());
                return new ArgoDeploymentInfo(deployResponse.getResourceName(), deployResponse.getRevisionHash());
            }
        }
    }

    @Override
    public void rollbackArgoApplication(Event event, String pipelineRepoUrl, StepData stepData, String argoApplicationName, String argoRevisionHash) {
        var deployResponse = clientWrapperApi.rollback(stepData.getClusterName(), event.getOrgName(), argoApplicationName, argoRevisionHash);
        log.info("Rolling back Argo application {}...", deployResponse.getResourceName());
        if (!deployResponse.getSuccess()) {
            var message = "Rolling back the Argo application failed.";
            log.error(message);
            throw new AtlasRetryableError(message);
        }
        var watchRequest = new WatchRequest(event.getTeamName(), event.getPipelineName(), stepData.getName(), WATCH_ARGO_APPLICATION_KEY, deployResponse.getResourceName(), deployResponse.getApplicationNamespace());
        clientWrapperApi.watchApplication(stepData.getClusterName(), event.getOrgName(), watchRequest);
        log.info("Watching rolled back Argo application {}", deployResponse.getResourceName());
    }

    @Override
    public void triggerStateRemediation(Event event, String pipelineRepoUrl, StepData stepData, String argoApplicationName, String argoRevisionHash, List<ResourceGvk> resourceStatuses) {
        var syncRequestPayload = new ResourcesGvkRequest(resourceStatuses);
        clientWrapperApi.selectiveSyncForArgoApp(
                stepData.getClusterName(),
                event.getOrgName(),
                event.getTeamName(),
                event.getPipelineName(),
                stepData.getName(),
                argoApplicationName,
                argoRevisionHash,
                syncRequestPayload
        );
    }

    @Override
    public boolean rollbackInPipelineExists(Event event, PipelineData pipelineData, String stepName) {
        var matchingSteps = metadataHandler.findAllStepsWithSameArgoRepoSrc(event, pipelineData, stepName);
        for (var step : matchingSteps) {
            var latestDeploymentLog = deploymentLogHandler.getLatestDeploymentLog(event, step);
            if (latestDeploymentLog != null
                    && (latestDeploymentLog.getUniqueVersionInstance() > 0 || latestDeploymentLog.getRollbackUniqueVersionNumber() != null)) {
                return true;
            }
        }
        return false;
    }
}
