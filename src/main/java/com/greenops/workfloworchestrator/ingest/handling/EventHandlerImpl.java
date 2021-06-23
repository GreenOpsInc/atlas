package com.greenops.workfloworchestrator.ingest.handling;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.workfloworchestrator.datamodel.event.ClientCompletionEvent;
import com.greenops.workfloworchestrator.datamodel.event.Event;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.PipelineData;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.StepData;
import com.greenops.workfloworchestrator.datamodel.requests.GetFileRequest;
import com.greenops.workfloworchestrator.ingest.apiclient.clientwrapper.ClientWrapperApi;
import com.greenops.workfloworchestrator.ingest.apiclient.reposerver.RepoManagerApi;
import com.greenops.workfloworchestrator.ingest.dbclient.DbClient;
import com.greenops.workfloworchestrator.ingest.dbclient.DbKey;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Component;

@Slf4j
@Component
public class EventHandlerImpl implements EventHandler {

    private static final String PIPELINE_FILE_NAME = "pipeline.yaml";
    private static final String NO_OP = "NO_OP_FOR_NOW";
    private static final String ARGO = "argo";

    private ClientWrapperApi clientWrapperApi;
    private RepoManagerApi repoManagerApi;
    private DbClient dbClient;
    private ObjectMapper yamlObjectMapper;
    private ObjectMapper objectMapper;

    @Autowired
    EventHandlerImpl(ClientWrapperApi clientWrapperApi,
                     RepoManagerApi repoManagerApi,
                     DbClient dbClient,
                     @Qualifier("yamlObjectMapper") ObjectMapper yamlObjectMapper,
                     @Qualifier("objectMapper") ObjectMapper objectMapper) {
        this.clientWrapperApi = clientWrapperApi;
        this.repoManagerApi = repoManagerApi;
        this.dbClient = dbClient;
        this.yamlObjectMapper = yamlObjectMapper;
        this.objectMapper = objectMapper;
    }

    @Override
    public boolean handleEvent(Event event) {
        if (event instanceof ClientCompletionEvent) {
            var teamSchema = dbClient.fetchTeamSchema(DbKey.makeDbTeamKey(event.getOrgName(), event.getTeamName()));
            if (teamSchema == null) return false;
            var gitRepoUrl = teamSchema.getPipelineSchema(event.getPipelineName()).getGitRepoSchema().getGitRepo();
            var getFileRequest = new GetFileRequest(gitRepoUrl, PIPELINE_FILE_NAME);
            try {
                var pipelineData = objectMapper.readValue(
                        objectMapper.writeValueAsString(
                                yamlObjectMapper.readValue(repoManagerApi.getFileFromRepo(getFileRequest, event.getOrgName(), event.getTeamName()), Object.class)
                        ),
                        PipelineData.class);
                var childrenSteps= pipelineData.getChildrenSteps(event.getStepName());
                for (var stepName : childrenSteps) {
                    var step = pipelineData.getStep(stepName);
                    if (areParentStepsComplete(stepName)) {
                        if (!deployStep(event.getPipelineName(), gitRepoUrl, step, event)) return false;
                    }
                }
            } catch (JsonProcessingException e) {
                log.error("Could not parse YAML pipeline data file", e);
            }
        }
        return true;
    }

    private boolean deployStep(String pipelineName, String pipelineRepoUrl, StepData stepData, Event event) {
        //TODO: Run before tests
        if (stepData.getOtherDeploymentsPath() != null) {
            var getFileRequest = new GetFileRequest(pipelineRepoUrl, stepData.getOtherDeploymentsPath());
            var otherDeploymentsConfig = repoManagerApi.getFileFromRepo(getFileRequest, event.getOrgName(), event.getTeamName());
            var deployResponse = clientWrapperApi.deploy(ARGO, NO_OP, NO_OP, otherDeploymentsConfig);
            if (!deployResponse.getSuccess()) {
                log.error("Deploying other resources failed.");
                return false;
            }
        }
        if (stepData.getArgoApplicationPath() != null) {
            var getFileRequest = new GetFileRequest(pipelineRepoUrl, stepData.getArgoApplicationPath());
            var argoApplicationConfig = repoManagerApi.getFileFromRepo(getFileRequest, event.getOrgName(), event.getTeamName());
            var deployResponse = clientWrapperApi.deploy(ARGO, NO_OP, NO_OP, argoApplicationConfig);
            log.info("Deploying application {}...", stepData.getArgoApplication());
            if (!deployResponse.getSuccess()) {
                log.error("Deploying the application failed.");
                return false;
            } else {
                var watching = clientWrapperApi.watchApplication(event.getOrgName(), event.getTeamName(), pipelineName, stepData.getName(), deployResponse.getApplicationNamespace(), stepData.getArgoApplication());
                if (!watching) return false;
                log.info("Watching application {}", stepData.getArgoApplication());
            }
        } else {
            //TODO: Expectation is argo application has already been created. Not currently supported.
        }
        //TODO: Audit log
        //TODO: Run after tests
        return true;
    }

    private boolean areParentStepsComplete(String stepName) {
        //TODO: Check redis and implement the flow for ensuring the parent steps have all been completed
        return true;
    }
}
