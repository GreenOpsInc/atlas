package com.greenops.workfloworchestrator.ingest.handling;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.util.datamodel.clientmessages.ResourceGvk;
import com.greenops.util.datamodel.clientmessages.ResourcesGvkRequest;
import com.greenops.util.datamodel.event.Event;
import com.greenops.util.datamodel.git.GitRepoSchemaInfo;
import com.greenops.util.datamodel.request.GetFileRequest;
import com.greenops.util.error.AtlasNonRetryableError;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.PipelineData;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.StepData;
import com.greenops.workfloworchestrator.ingest.apiclient.clientwrapper.ClientRequestQueue;
import com.greenops.workfloworchestrator.ingest.apiclient.reposerver.RepoManagerApi;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Component;

import java.util.List;

import static com.greenops.workfloworchestrator.datamodel.pipelinedata.StepData.ROOT_STEP_NAME;
import static com.greenops.workfloworchestrator.datamodel.pipelinedata.Test.DEFAULT_NAMESPACE;
import static com.greenops.workfloworchestrator.ingest.handling.EventHandlerImpl.WATCH_ARGO_APPLICATION_KEY;

@Slf4j
@Component
public class DeploymentHandlerImpl implements DeploymentHandler {

    private RepoManagerApi repoManagerApi;
    private ClientRequestQueue clientRequestQueue;
    private MetadataHandler metadataHandler;
    private DeploymentLogHandler deploymentLogHandler;
    private ObjectMapper yamlObjectMapper;

    @Autowired
    DeploymentHandlerImpl(RepoManagerApi repoManagerApi,
                          ClientRequestQueue clientRequestQueue,
                          MetadataHandler metadataHandler,
                          DeploymentLogHandler deploymentLogHandler,
                          @Qualifier("yamlObjectMapper") ObjectMapper yamlObjectMapper) {
        this.repoManagerApi = repoManagerApi;
        this.clientRequestQueue = clientRequestQueue;
        this.metadataHandler = metadataHandler;
        this.deploymentLogHandler = deploymentLogHandler;
        this.yamlObjectMapper = yamlObjectMapper;
    }

    @Override
    public void deleteApplicationInfrastructure(Event event, GitRepoSchemaInfo gitRepoSchemaInfo, StepData stepData, String gitCommitHash) {
        if (stepData.getOtherDeploymentsPath() != null) {
            var getFileRequest = new GetFileRequest(gitRepoSchemaInfo, stepData.getOtherDeploymentsPath(), gitCommitHash);
            var otherDeploymentsConfig = repoManagerApi.getFileFromRepo(getFileRequest, event.getOrgName(), event.getTeamName());
            log.info("Deleting old application infrastructure...");
            if (otherDeploymentsConfig.isBlank()) return;
            clientRequestQueue.deleteByConfig(
                    stepData.getClusterName(),
                    event.getOrgName(),
                    event.getTeamName(),
                    event.getPipelineName(),
                    event.getPipelineUvn(),
                    event.getStepName(),
                    getStepNamespace(event, repoManagerApi, yamlObjectMapper, stepData.getArgoApplicationPath(), gitRepoSchemaInfo, gitCommitHash),
                    ClientRequestQueue.DELETE_KUBERNETES_REQUEST,
                    otherDeploymentsConfig
            );
        }
    }

    @Override
    public void deployApplicationInfrastructure(Event event, GitRepoSchemaInfo gitRepoSchemaInfo, StepData stepData, String gitCommitHash) {
        if (stepData.getOtherDeploymentsPath() != null) {
            var getFileRequest = new GetFileRequest(gitRepoSchemaInfo, stepData.getOtherDeploymentsPath(), gitCommitHash);
            var otherDeploymentsConfig = repoManagerApi.getFileFromRepo(getFileRequest, event.getOrgName(), event.getTeamName());
            log.info("Deploying new application infrastructure...");
            if (otherDeploymentsConfig.isBlank()) return;
            clientRequestQueue.deploy(
                    stepData.getClusterName(),
                    event.getOrgName(),
                    event.getTeamName(),
                    event.getPipelineName(),
                    event.getPipelineUvn(),
                    stepData.getName(),
                    getStepNamespace(event, repoManagerApi, yamlObjectMapper, stepData.getArgoApplicationPath(), gitRepoSchemaInfo, gitCommitHash),
                    ClientRequestQueue.RESPONSE_EVENT_APPLICATION_INFRA,
                    ClientRequestQueue.DEPLOY_KUBERNETES_REQUEST,
                    ClientRequestQueue.LATEST_REVISION,
                    otherDeploymentsConfig
            );
        }
    }

    @Override
    public void deployArgoApplication(Event event, GitRepoSchemaInfo gitRepoSchemaInfo, PipelineData pipelineData, String stepName, String argoRevisionHash, String gitCommitHash) {
        var stepData = pipelineData.getStep(stepName);
        if (stepData.getArgoApplicationPath() != null) {
            var getFileRequest = new GetFileRequest(gitRepoSchemaInfo, stepData.getArgoApplicationPath(), gitCommitHash);
            var argoApplicationConfig = repoManagerApi.getFileFromRepo(getFileRequest, event.getOrgName(), event.getTeamName());
            metadataHandler.assertArgoRepoMetadataExists(event, stepData.getName(), argoApplicationConfig);
            var pipelineLockRevisionHash = pipelineData.isArgoVersionLock() ? metadataHandler.getPipelineLockRevisionHash(event, pipelineData, stepData.getName()) : null;
            pipelineLockRevisionHash = pipelineLockRevisionHash != null ? pipelineLockRevisionHash : ClientRequestQueue.LATEST_REVISION;
            clientRequestQueue.deployAndWatch(
                    stepData.getClusterName(),
                    event.getOrgName(),
                    event.getTeamName(),
                    event.getPipelineName(),
                    event.getPipelineUvn(),
                    event.getStepName(),
                    getStepNamespace(event, repoManagerApi, yamlObjectMapper, stepData.getArgoApplicationPath(), gitRepoSchemaInfo, gitCommitHash),
                    ClientRequestQueue.DEPLOY_ARGO_REQUEST,
                    pipelineLockRevisionHash,
                    argoApplicationConfig,
                    WATCH_ARGO_APPLICATION_KEY,
                    -1
            );
            log.info("Deploying and watching Argo application...");
        } else { //stepData.getArgoApplication() != null
            //Disabling deployin argo app by name for now
//            var deployResponse = clientWrapperApi.deployArgoAppByName(stepData.getClusterName(), event.getOrgName(), stepData.getArgoApplication());
//            log.info("Syncing the Argo application {}...", deployResponse.getResourceName());
//            if (!deployResponse.getSuccess()) {
//                var message = "Syncing the Argo application failed.";
//                log.error(message);
//                throw new AtlasRetryableError(message);
//            } else {
//                var watchRequest = new WatchRequest(event.getTeamName(), event.getPipelineName(), stepData.getName(), WATCH_ARGO_APPLICATION_KEY, deployResponse.getResourceName(), deployResponse.getApplicationNamespace());
//                clientWrapperApi.watchApplication(stepData.getClusterName(), event.getOrgName(), watchRequest);
//                log.info("Watching Argo application {}", deployResponse.getResourceName());
//                return new ArgoDeploymentInfo(deployResponse.getResourceName(), deployResponse.getRevisionHash());
//            }
        }
    }

    @Override
    public void rollbackArgoApplication(Event event, GitRepoSchemaInfo gitRepoSchemaInfo, StepData stepData, String argoApplicationName, String argoRevisionHash) {
        clientRequestQueue.rollbackAndWatch(
                stepData.getClusterName(),
                event.getOrgName(),
                event.getTeamName(),
                event.getPipelineName(),
                event.getPipelineUvn(),
                event.getStepName(),
                getStepNamespace(event, repoManagerApi, yamlObjectMapper, stepData.getArgoApplicationPath(), gitRepoSchemaInfo, argoRevisionHash),
                argoApplicationName,
                argoRevisionHash,
                WATCH_ARGO_APPLICATION_KEY
        );
        log.info("Rolling back and watching Argo application...");
    }

    @Override
    public void triggerStateRemediation(Event event, GitRepoSchemaInfo gitRepoSchemaInfo, StepData stepData, String argoApplicationName, String argoRevisionHash, List<ResourceGvk> resourceStatuses) {
        var syncRequestPayload = new ResourcesGvkRequest(resourceStatuses);
        clientRequestQueue.selectiveSyncArgoApplication(
                stepData.getClusterName(),
                event.getOrgName(),
                event.getTeamName(),
                event.getPipelineName(),
                event.getPipelineUvn(),
                stepData.getName(),
                getStepNamespace(event, repoManagerApi, yamlObjectMapper, stepData.getArgoApplicationPath(), gitRepoSchemaInfo, argoRevisionHash),
                argoRevisionHash,
                syncRequestPayload,
                argoApplicationName
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

    public static String getStepNamespace(Event event, RepoManagerApi repoManagerApi, ObjectMapper yamlObjectMapper, String argoApplicationPath, GitRepoSchemaInfo gitRepoSchemaInfo, String gitCommitHash) {
        if (event.getStepName().isEmpty() || event.getStepName().equals(ROOT_STEP_NAME)) {
            throw new AtlasNonRetryableError("Could not find a namespace associated with the event");
        }

        if (argoApplicationPath == null || argoApplicationPath.isEmpty()) {
            return DEFAULT_NAMESPACE;
        }
        var getFileRequest = new GetFileRequest(gitRepoSchemaInfo, argoApplicationPath, gitCommitHash);
        var argoAppPayload = repoManagerApi.getFileFromRepo(getFileRequest, event.getOrgName(), event.getTeamName());
        String namespace;
        try {
            var sourceJsonNode = yamlObjectMapper.readTree(argoAppPayload).get("spec").get("destination");
            namespace = sourceJsonNode.get("namespace").asText(null);
        } catch (JsonProcessingException e) {
            throw new AtlasNonRetryableError("Argo app configuration cannot be parsed");
        }
        if (namespace == null) {
            throw new AtlasNonRetryableError("Could not find a namespace associated with the event");
        }
        return namespace;
    }
}
