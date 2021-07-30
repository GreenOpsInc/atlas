package com.greenops.workfloworchestrator.ingest.handling;

import com.greenops.workfloworchestrator.datamodel.event.Event;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.StepData;
import com.greenops.workfloworchestrator.datamodel.requests.GetFileRequest;
import com.greenops.workfloworchestrator.datamodel.requests.WatchRequest;
import com.greenops.workfloworchestrator.error.AtlasRetryableError;
import com.greenops.workfloworchestrator.ingest.apiclient.clientwrapper.ClientWrapperApi;
import com.greenops.workfloworchestrator.ingest.apiclient.reposerver.RepoManagerApi;
import com.greenops.workfloworchestrator.ingest.handling.util.deployment.ArgoDeploymentInfo;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import static com.greenops.workfloworchestrator.ingest.handling.EventHandlerImpl.WATCH_ARGO_APPLICATION_KEY;
import static com.greenops.workfloworchestrator.ingest.handling.util.deployment.ArgoDeploymentInfo.NO_OP_ARGO_DEPLOYMENT;

@Slf4j
@Component
public class DeploymentHandlerImpl implements DeploymentHandler {

    private RepoManagerApi repoManagerApi;
    private ClientWrapperApi clientWrapperApi;

    @Autowired
    DeploymentHandlerImpl(RepoManagerApi repoManagerApi, ClientWrapperApi clientWrapperApi) {
        this.repoManagerApi = repoManagerApi;
        this.clientWrapperApi = clientWrapperApi;
    }

    @Override
    public void deleteApplicationInfrastructure(Event event, String pipelineRepoUrl, StepData stepData, String gitCommitHash) {
        if (stepData.getOtherDeploymentsPath() != null) {
            var getFileRequest = new GetFileRequest(pipelineRepoUrl, stepData.getOtherDeploymentsPath(), gitCommitHash);
            var otherDeploymentsConfig = repoManagerApi.getFileFromRepo(getFileRequest, event.getOrgName(), event.getTeamName());
            log.info("Deleting old application infrastructure...");
            for (var deploymentConfig : otherDeploymentsConfig.split("---")) {
                if (deploymentConfig.isBlank()) continue;
                clientWrapperApi.delete(event.getOrgName(), ClientWrapperApi.DELETE_KUBERNETES_REQUEST, deploymentConfig);
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
                var deployResponse = clientWrapperApi.deploy(event.getOrgName(), ClientWrapperApi.DEPLOY_KUBERNETES_REQUEST, deploymentConfig);
                if (!deployResponse.getSuccess()) {
                    var message = "Deploying other resources failed.";
                    log.error(message);
                    throw new AtlasRetryableError(message);
                }
            }
        }
    }

    @Override
    public ArgoDeploymentInfo deployArgoApplication(Event event, String pipelineRepoUrl, StepData stepData, String gitCommitHash) {
        if (stepData.getArgoApplicationPath() != null) {
            var getFileRequest = new GetFileRequest(pipelineRepoUrl, stepData.getArgoApplicationPath(), gitCommitHash);
            var argoApplicationConfig = repoManagerApi.getFileFromRepo(getFileRequest, event.getOrgName(), event.getTeamName());
            //TODO: The splitting of the config file should eventually be done on the client side
            for (var applicationConfig : argoApplicationConfig.split("---")) {
                var deployResponse = clientWrapperApi.deploy(event.getOrgName(), ClientWrapperApi.DEPLOY_ARGO_REQUEST, applicationConfig);
                log.info("Deploying Argo application {}...", deployResponse.getResourceName());
                if (!deployResponse.getSuccess()) {
                    var message = "Deploying the Argo application failed.";
                    log.error(message);
                    throw new AtlasRetryableError(message);
                } else {
                    var watchRequest = new WatchRequest(event.getTeamName(), event.getPipelineName(), stepData.getName(), WATCH_ARGO_APPLICATION_KEY, deployResponse.getResourceName(), deployResponse.getApplicationNamespace());
                    clientWrapperApi.watchApplication(event.getOrgName(), watchRequest);
                    log.info("Watching Argo application {}", deployResponse.getResourceName());
                    return new ArgoDeploymentInfo(deployResponse.getResourceName());
                }
            }
        } else { //stepData.getArgoApplication() != null
            var deployResponse = clientWrapperApi.deployArgoAppByName(event.getOrgName(), stepData.getArgoApplication());
            log.info("Syncing the Argo application {}...", deployResponse.getResourceName());
            if (!deployResponse.getSuccess()) {
                var message = "Syncing the Argo application failed.";
                log.error(message);
                throw new AtlasRetryableError(message);
            } else {
                var watchRequest = new WatchRequest(event.getTeamName(), event.getPipelineName(), stepData.getName(), WATCH_ARGO_APPLICATION_KEY, deployResponse.getResourceName(), deployResponse.getApplicationNamespace());
                clientWrapperApi.watchApplication(event.getOrgName(), watchRequest);
                log.info("Watching Argo application {}", deployResponse.getResourceName());
                return new ArgoDeploymentInfo(deployResponse.getResourceName());
            }
        }
        return NO_OP_ARGO_DEPLOYMENT;
    }

    @Override
    public void rollbackArgoApplication(Event event, String pipelineRepoUrl, StepData stepData, String argoApplicationName, int argoRevisionId) {
        var deployResponse = clientWrapperApi.rollback(event.getOrgName(), argoApplicationName, argoRevisionId);
        log.info("Rolling back Argo application {}...", deployResponse.getResourceName());
        if (!deployResponse.getSuccess()) {
            var message = "Rolling back the Argo application failed.";
            log.error(message);
            throw new AtlasRetryableError(message);
        }
        var watchRequest = new WatchRequest(event.getTeamName(), event.getPipelineName(), stepData.getName(), WATCH_ARGO_APPLICATION_KEY, deployResponse.getResourceName(), deployResponse.getApplicationNamespace());
        clientWrapperApi.watchApplication(event.getOrgName(), watchRequest);
        log.info("Watching rolled back Argo application {}", deployResponse.getResourceName());
    }
}
