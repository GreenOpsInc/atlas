package com.greenops.workfloworchestrator.ingest.handling;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.workfloworchestrator.datamodel.event.ClientCompletionEvent;
import com.greenops.workfloworchestrator.datamodel.event.Event;
import com.greenops.workfloworchestrator.datamodel.event.TestCompletionEvent;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.PipelineData;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.StepData;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.Test;
import com.greenops.workfloworchestrator.datamodel.pipelineschema.TeamSchema;
import com.greenops.workfloworchestrator.datamodel.requests.GetFileRequest;
import com.greenops.workfloworchestrator.ingest.apiclient.reposerver.RepoManagerApi;
import com.greenops.workfloworchestrator.ingest.dbclient.DbClient;
import com.greenops.workfloworchestrator.ingest.dbclient.DbKey;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Component;

import java.util.stream.Collectors;

import static com.greenops.workfloworchestrator.datamodel.pipelinedata.StepData.ROOT_STEP_NAME;
import static com.greenops.workfloworchestrator.datamodel.pipelinedata.StepData.createRootStep;
import static com.greenops.workfloworchestrator.ingest.handling.ClientKey.getTestNumberFromTestKey;
import static com.greenops.workfloworchestrator.ingest.handling.util.deployment.ArgoDeploymentInfo.NO_OP_ARGO_DEPLOYMENT;

@Slf4j
@Component
public class EventHandlerImpl implements EventHandler {

    static final String WATCH_ARGO_APPLICATION_KEY = "WatchArgoApplicationKey";
    static final String WATCH_TEST_KEY = "WatchTestKey";
    static final String PIPELINE_FILE_NAME = "pipeline.yaml";

    private RepoManagerApi repoManagerApi;
    private DbClient dbClient;
    private DeploymentHandler deploymentHandler;
    private TestHandler testHandler;
    private DeploymentLogHandler deploymentLogHandler;
    private ObjectMapper yamlObjectMapper;
    private ObjectMapper objectMapper;

    @Autowired
    EventHandlerImpl(RepoManagerApi repoManagerApi,
                     DbClient dbClient,
                     DeploymentHandler deploymentHandler,
                     TestHandler testHandler,
                     DeploymentLogHandler deploymentLogHandler,
                     @Qualifier("yamlObjectMapper") ObjectMapper yamlObjectMapper,
                     @Qualifier("objectMapper") ObjectMapper objectMapper) {
        this.repoManagerApi = repoManagerApi;
        this.dbClient = dbClient;
        this.deploymentHandler = deploymentHandler;
        this.testHandler = testHandler;
        this.deploymentLogHandler = deploymentLogHandler;
        this.yamlObjectMapper = yamlObjectMapper;
        this.objectMapper = objectMapper;
    }

    @Override
    public boolean handleEvent(Event event) {
        var teamSchema = fetchTeamSchema(event);
        if (teamSchema == null) return false;
        var pipelineData = fetchPipelineData(event, teamSchema);
        if (pipelineData == null) return false;
        var gitRepoUrl = teamSchema.getPipelineSchema(event.getPipelineName()).getGitRepoSchema().getGitRepo();
        if (event instanceof ClientCompletionEvent) {
            return handleClientCompletionEvent(pipelineData, gitRepoUrl, (ClientCompletionEvent) event);
        } else if (event instanceof TestCompletionEvent) {
            return handleTestCompletion(pipelineData, gitRepoUrl, (TestCompletionEvent) event);
        }
        return true;
    }

    private boolean handleClientCompletionEvent(PipelineData pipelineData, String pipelineRepoUrl, ClientCompletionEvent event) {
        if (!deploymentLogHandler.markDeploymentSuccessful(event, event.getStepName())) {
            return false;
        }

        //TODO: How do we decide whether a deployment was unsucessful?

        if (event.getStepName().equals(ROOT_STEP_NAME)) {
            return triggerNextSteps(pipelineData, createRootStep(), pipelineRepoUrl, event);
        }
        var step = pipelineData.getStep(event.getStepName());
        var afterTestsExist = step.getTests().stream().anyMatch(test -> !test.shouldExecuteBefore());
        if (afterTestsExist) {
            return testHandler.triggerTest(pipelineRepoUrl, step, false, event);
        } else {
            return triggerNextSteps(pipelineData, step, pipelineRepoUrl, event);
        }
    }

    private boolean handleTestCompletion(PipelineData pipelineData, String pipelineRepoUrl, TestCompletionEvent event) {
        var step = pipelineData.getStep(event.getStepName());
        if (!event.getSuccessful()) {
            deploymentLogHandler.markStepFailedWithBrokenTest(event, event.getStepName(), event.getTestName(), event.getLog());
            if (step.getRollback()) return rollback(pipelineData, pipelineRepoUrl, event);
            return true;
        }

        var completedTestNumber = getTestNumberFromTestKey(event.getTestName());
        if (completedTestNumber < 0 || step.getTests().size() <= completedTestNumber) {
            log.info("Malformed test key or tests have changed");
            return false;
        }
        var completedTest = step.getTests().get(completedTestNumber);
        var tests = step.getTests().stream().filter(test -> test.shouldExecuteBefore() == completedTest.shouldExecuteBefore()).collect(Collectors.toList());

        if (completedTest.shouldExecuteBefore() && completedTestNumber == tests.size() - 1) {
            return deploy(pipelineRepoUrl, step, event);
        } else if (!completedTest.shouldExecuteBefore() && completedTestNumber == tests.size() - 1) {
            return triggerNextSteps(pipelineData, step, pipelineRepoUrl, event);
        } else if (completedTestNumber < tests.size()) {
            return testHandler.createAndRunTest(step.getName(), pipelineRepoUrl, step.getTests().get(completedTestNumber + 1), completedTestNumber + 1, event);
        } else {
            //This case should never be happening...log and see what the edge case is
            log.info("EDGE CASE: {}, {}", completedTest.shouldExecuteBefore(), completedTestNumber == step.getTests().size() - 1);
        }
        return true;
    }

    private boolean triggerNextSteps(PipelineData pipelineData, StepData step, String pipelineRepoUrl, Event event) {
        if (!deploymentLogHandler.markStepSuccessful(event, event.getStepName())) {
            return false;
        }

        var childrenSteps = pipelineData.getChildrenSteps(step.getName());
        for (var stepName : childrenSteps) {
            var nextStep = pipelineData.getStep(stepName);
            if (deploymentLogHandler.areParentStepsComplete(stepName)) {
                if (!deploymentLogHandler.initializeNewStepLog(
                        event,
                        nextStep.getName(),
                        repoManagerApi.getCurrentPipelineCommitHash(pipelineRepoUrl, event.getOrgName(), event.getTeamName())
                )) {
                    return false;
                }
                if (!triggerStep(event.getPipelineName(), pipelineRepoUrl, nextStep, event)) return false;
            }
        }
        return true;
    }

    private boolean triggerStep(String pipelineName, String pipelineRepoUrl, StepData stepData, Event event) {
        var beforeTestsExist = stepData.getTests().stream().anyMatch(Test::shouldExecuteBefore);
        if (beforeTestsExist) {
            return testHandler.triggerTest(pipelineRepoUrl, stepData, true, event);
        }

        if (stepData.getOtherDeploymentsPath() != null || stepData.getArgoApplicationPath() != null) {
            return deploy(pipelineRepoUrl, stepData, event);
        }

        var afterTestsExist = stepData.getTests().stream().anyMatch(test -> !test.shouldExecuteBefore());
        if (afterTestsExist) {
            return testHandler.triggerTest(pipelineRepoUrl, stepData, false, event);
        }
        return true;
    }

    private boolean deploy(String pipelineRepoUrl, StepData stepData, Event event) {
        if (!deploymentHandler.deployApplicationInfrastructure(event, pipelineRepoUrl, stepData)) return false;

        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepData.getName());
        var deploymentLog = dbClient.fetchLatestLog(logKey);
        if (deploymentLog == null) return false;

        var argoDeploymentInfo = NO_OP_ARGO_DEPLOYMENT;
        if ((stepData.getArgoApplicationPath() != null || stepData.getArgoApplication() != null) && deploymentLog.getUniqueVersionInstance() > 0) {
            return deploymentHandler.rollbackArgoApplication(event, pipelineRepoUrl, stepData, deploymentLog.getArgoApplicationName(), deploymentLog.getArgoRevisionId());
        } else if (stepData.getArgoApplicationPath() != null || stepData.getArgoApplication() != null) {
            argoDeploymentInfo = deploymentHandler.deployArgoApplication(event, pipelineRepoUrl, stepData);
        }

        //Audit log updates
        return deploymentLogHandler.updateStepDeploymentLog(event, stepData.getName(), argoDeploymentInfo.getArgoApplicationName(), argoDeploymentInfo.getArgoRevisionId());
    }

    private boolean rollback(PipelineData pipelineData, String pipelineRepoUrl, TestCompletionEvent event) {
        var gitCommitVersion = deploymentLogHandler.makeRollbackDeploymentLog(event, event.getStepName());
        if (gitCommitVersion == null) {
            return false;
        } else if (gitCommitVersion.isEmpty()) {
            //Means there is no stable version that can be found.
            return true;
        }
        if (!repoManagerApi.resetRepoVersion(gitCommitVersion, pipelineRepoUrl, event.getOrgName(), event.getTeamName())) {
            return false;
        }
        return triggerStep(event.getPipelineName(), pipelineRepoUrl, pipelineData.getStep(event.getStepName()), event);
    }

    private TeamSchema fetchTeamSchema(Event event) {
        return dbClient.fetchTeamSchema(DbKey.makeDbTeamKey(event.getOrgName(), event.getTeamName()));
    }

    private PipelineData fetchPipelineData(Event event, TeamSchema teamSchema) {
        var gitRepoUrl = teamSchema.getPipelineSchema(event.getPipelineName()).getGitRepoSchema().getGitRepo();
        var getFileRequest = new GetFileRequest(gitRepoUrl, PIPELINE_FILE_NAME);
        try {
            return objectMapper.readValue(
                    objectMapper.writeValueAsString(
                            yamlObjectMapper.readValue(repoManagerApi.getFileFromRepo(getFileRequest, event.getOrgName(), event.getTeamName()), Object.class)
                    ),
                    PipelineData.class);
        } catch (JsonProcessingException e) {
            log.error("Could not parse YAML pipeline data file", e);
        }
        return null;
    }
}
