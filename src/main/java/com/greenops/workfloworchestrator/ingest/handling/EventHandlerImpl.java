package com.greenops.workfloworchestrator.ingest.handling;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.workfloworchestrator.datamodel.event.*;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.PipelineData;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.StepData;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.Test;
import com.greenops.workfloworchestrator.datamodel.pipelineschema.TeamSchema;
import com.greenops.workfloworchestrator.datamodel.requests.GetFileRequest;
import com.greenops.workfloworchestrator.error.AtlasNonRetryableError;
import com.greenops.workfloworchestrator.ingest.apiclient.reposerver.RepoManagerApi;
import com.greenops.workfloworchestrator.ingest.dbclient.DbClient;
import com.greenops.workfloworchestrator.ingest.dbclient.DbKey;
import com.greenops.workfloworchestrator.ingest.kafka.KafkaClient;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Component;

import java.util.stream.Collectors;

import static com.greenops.workfloworchestrator.datamodel.event.ClientCompletionEvent.*;
import static com.greenops.workfloworchestrator.datamodel.pipelinedata.StepData.ROOT_STEP_NAME;
import static com.greenops.workfloworchestrator.datamodel.pipelinedata.StepData.createRootStep;
import static com.greenops.workfloworchestrator.ingest.apiclient.reposerver.RepoManagerApi.ROOT_COMMIT;
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
    private KafkaClient kafkaClient;
    private ObjectMapper yamlObjectMapper;
    private ObjectMapper objectMapper;

    @Autowired
    EventHandlerImpl(RepoManagerApi repoManagerApi,
                     DbClient dbClient,
                     DeploymentHandler deploymentHandler,
                     TestHandler testHandler,
                     DeploymentLogHandler deploymentLogHandler,
                     KafkaClient kafkaClient,
                     @Qualifier("yamlObjectMapper") ObjectMapper yamlObjectMapper,
                     @Qualifier("objectMapper") ObjectMapper objectMapper) {
        this.repoManagerApi = repoManagerApi;
        this.dbClient = dbClient;
        this.deploymentHandler = deploymentHandler;
        this.testHandler = testHandler;
        this.deploymentLogHandler = deploymentLogHandler;
        this.kafkaClient = kafkaClient;
        this.yamlObjectMapper = yamlObjectMapper;
        this.objectMapper = objectMapper;
    }

    @Override
    public void handleEvent(Event event) {
        var teamSchema = fetchTeamSchema(event);
        if (teamSchema == null) throw new AtlasNonRetryableError("The team doesn't exist");
        //TODO: When the trigger step event is added, gitCommit should be ROOT_COMMIT when the event is a triggerstepevent
        var gitCommit = event.getStepName().equals(ROOT_STEP_NAME) ? ROOT_COMMIT : deploymentLogHandler.getCurrentGitCommitHash(event, event.getStepName());
        var pipelineData = fetchPipelineData(event, teamSchema, gitCommit);
        if (pipelineData == null) throw new AtlasNonRetryableError("The pipeline doesn't exist");
        var gitRepoUrl = teamSchema.getPipelineSchema(event.getPipelineName()).getGitRepoSchema().getGitRepo();
        if (event instanceof ClientCompletionEvent) {
            handleClientCompletionEvent(pipelineData, gitRepoUrl, (ClientCompletionEvent) event);
        } else if (event instanceof TestCompletionEvent) {
            handleTestCompletion(pipelineData, gitRepoUrl, (TestCompletionEvent) event);
        } else if (event instanceof ApplicationInfraTriggerEvent) {
            handleApplicationInfraTrigger(teamSchema, pipelineData, gitRepoUrl, (ApplicationInfraTriggerEvent) event);
        } else if (event instanceof ApplicationInfraCompletionEvent) {
            handleApplicationInfraCompletion(gitRepoUrl, pipelineData, (ApplicationInfraCompletionEvent) event);
        }
    }

    private void handleClientCompletionEvent(PipelineData pipelineData, String pipelineRepoUrl, ClientCompletionEvent event) {
        if (event.getHealthStatus().equals(PROGRESSING)) {
            return;
        }

        var step = pipelineData.getStep(event.getStepName());
        if (event.getHealthStatus().equals(DEGRADED) || event.getHealthStatus().equals(UNKNOWN)) {
            deploymentLogHandler.markStepFailedWithFailedDeployment(event, event.getStepName(), event.getRevisionId());
            if (step.getRollback()) rollback(pipelineData, pipelineRepoUrl, event);
            return;
        }
        //TODO: How do we handle the remaining sync/health statuses? We should be retriggering syncs, waiting for status updates (?), etc

        deploymentLogHandler.markDeploymentSuccessful(event, event.getStepName(), event.getRevisionId());

        if (event.getStepName().equals(ROOT_STEP_NAME)) {
            triggerNextSteps(pipelineData, createRootStep(), pipelineRepoUrl, event);
            return;
        }

        var afterTestsExist = step.getTests().stream().anyMatch(test -> !test.shouldExecuteBefore());
        if (afterTestsExist) {
            testHandler.triggerTest(pipelineRepoUrl, step, false, deploymentLogHandler.getCurrentGitCommitHash(event, step.getName()), event);
        } else {
            triggerNextSteps(pipelineData, step, pipelineRepoUrl, event);
        }
    }

    private void handleTestCompletion(PipelineData pipelineData, String pipelineRepoUrl, TestCompletionEvent event) {
        var step = pipelineData.getStep(event.getStepName());
        if (!event.getSuccessful()) {
            deploymentLogHandler.markStepFailedWithBrokenTest(event, event.getStepName(), event.getTestName(), event.getLog());
            if (step.getRollback()) rollback(pipelineData, pipelineRepoUrl, event);
            return;
        }

        var completedTestNumber = event.getTestNumber();
        if (completedTestNumber < 0 || step.getTests().size() <= completedTestNumber) {
            log.info("Malformed test key or tests have changed. This event will be ignored.");
            return;
        }
        var completedTest = step.getTests().get(completedTestNumber);
        var tests = step.getTests().stream().filter(test -> test.shouldExecuteBefore() == completedTest.shouldExecuteBefore()).collect(Collectors.toList());

        if (completedTest.shouldExecuteBefore() && completedTestNumber == tests.size() - 1) {
            triggerAppInfraDeploy(step.getName(), event);
        } else if (!completedTest.shouldExecuteBefore() && completedTestNumber == tests.size() - 1) {
            triggerNextSteps(pipelineData, step, pipelineRepoUrl, event);
        } else if (completedTestNumber < tests.size()) {
            testHandler.createAndRunTest(
                    step.getName(),
                    pipelineRepoUrl,
                    step.getTests().get(completedTestNumber + 1),
                    completedTestNumber + 1,
                    deploymentLogHandler.getCurrentGitCommitHash(event, step.getName()),
                    event
            );
        } else {
            //This case should never be happening...log and see what the edge case is
            log.info("EDGE CASE: {}, {}", completedTest.shouldExecuteBefore(), completedTestNumber == step.getTests().size() - 1);
        }
    }

    private void handleApplicationInfraTrigger(TeamSchema teamSchema, PipelineData pipelineData, String pipelineRepoUrl, ApplicationInfraTriggerEvent event) {
        //Right now it is assumed that the step names do not change
        var stepData = pipelineData.getStep(event.getStepName());
        var oldGitCommitHash = stepData.getRollback()
                ? deploymentLogHandler.getLastSuccessfulStepGitCommitHash(event, event.getStepName())
                //TODO: Add deploymentInfraSuccessful variable in deployment log, and replace method below with one that checks for the deployment infra deploying successfully
                : deploymentLogHandler.getLastSuccessfulDeploymentGitCommitHash(event, event.getStepName());
        if (oldGitCommitHash != null) {
            var oldPipelineData = fetchPipelineData(event, teamSchema, oldGitCommitHash);
            var oldStepData = oldPipelineData.getStep(event.getStepName());
            deploymentHandler.deleteApplicationInfrastructure(event, pipelineRepoUrl, oldStepData, oldGitCommitHash);
        }
        deploymentHandler.deployApplicationInfrastructure(event, pipelineRepoUrl, pipelineData.getStep(event.getStepName()), deploymentLogHandler.getCurrentGitCommitHash(event, event.getStepName()));
        notifyAppInfraCompletion(event.getStepName(), event);
    }

    private void handleApplicationInfraCompletion(String pipelineRepoUrl, PipelineData pipelineData, ApplicationInfraCompletionEvent event) {
        var stepData = pipelineData.getStep(event.getStepName());
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepData.getName());
        var deploymentLog = dbClient.fetchLatestLog(logKey);

        var argoDeploymentInfo = NO_OP_ARGO_DEPLOYMENT;
        if ((stepData.getArgoApplicationPath() != null || stepData.getArgoApplication() != null) && deploymentLog.getUniqueVersionInstance() > 0) {
            deploymentHandler.rollbackArgoApplication(event, pipelineRepoUrl, stepData, deploymentLog.getArgoApplicationName(), deploymentLog.getArgoRevisionId());
            return;
        } else if (stepData.getArgoApplicationPath() != null || stepData.getArgoApplication() != null) {
            argoDeploymentInfo = deploymentHandler.deployArgoApplication(event, pipelineRepoUrl, stepData, deploymentLogHandler.getCurrentGitCommitHash(event, stepData.getName()));
        }

        //Audit log updates
        deploymentLogHandler.updateStepDeploymentLog(event, stepData.getName(), argoDeploymentInfo.getArgoApplicationName(), argoDeploymentInfo.getArgoRevisionId());
    }

    private void triggerNextSteps(PipelineData pipelineData, StepData step, String pipelineRepoUrl, Event event) {
        deploymentLogHandler.markStepSuccessful(event, event.getStepName());

        var childrenSteps = pipelineData.getChildrenSteps(step.getName());
        for (var stepName : childrenSteps) {
            var nextStep = pipelineData.getStep(stepName);
            var parentSteps = pipelineData.getParentSteps(stepName);
            if (deploymentLogHandler.areParentStepsComplete(event, parentSteps)) {
                deploymentLogHandler.initializeNewStepLog(
                        event,
                        nextStep.getName(),
                        repoManagerApi.getCurrentPipelineCommitHash(pipelineRepoUrl, event.getOrgName(), event.getTeamName())
                );
                triggerStep(event.getPipelineName(), pipelineRepoUrl, nextStep, event);
            }
        }
    }

    private void triggerStep(String pipelineName, String pipelineRepoUrl, StepData stepData, Event event) {
        var beforeTestsExist = stepData.getTests().stream().anyMatch(Test::shouldExecuteBefore);
        if (beforeTestsExist) {
            testHandler.triggerTest(pipelineRepoUrl, stepData, true, ROOT_COMMIT, event);
            return;
        }

        if (stepData.getOtherDeploymentsPath() != null || stepData.getArgoApplicationPath() != null) {
            triggerAppInfraDeploy(stepData.getName(), event);
            return;
        }

        var afterTestsExist = stepData.getTests().stream().anyMatch(test -> !test.shouldExecuteBefore());
        if (afterTestsExist) {
            testHandler.triggerTest(pipelineRepoUrl, stepData, false, ROOT_COMMIT, event);
            return;
        }
    }

    private void triggerAppInfraDeploy(String stepName, Event event) {
        kafkaClient.sendMessage(new ApplicationInfraTriggerEvent(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName));
    }

    private void notifyAppInfraCompletion(String stepName, Event event) {
        kafkaClient.sendMessage(new ApplicationInfraCompletionEvent(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName));
    }

    private void rollback(PipelineData pipelineData, String pipelineRepoUrl, Event event) {
        var gitCommitVersion = deploymentLogHandler.makeRollbackDeploymentLog(event, event.getStepName());
        if (gitCommitVersion.isEmpty()) {
            //Means there is no stable version that can be found.
            return;
        }
        repoManagerApi.resetRepoVersion(gitCommitVersion, pipelineRepoUrl, event.getOrgName(), event.getTeamName());
        triggerStep(event.getPipelineName(), pipelineRepoUrl, pipelineData.getStep(event.getStepName()), event);
    }

    private TeamSchema fetchTeamSchema(Event event) {
        return dbClient.fetchTeamSchema(DbKey.makeDbTeamKey(event.getOrgName(), event.getTeamName()));
    }

    private PipelineData fetchPipelineData(Event event, TeamSchema teamSchema, String gitCommitHash) {
        var gitRepoUrl = teamSchema.getPipelineSchema(event.getPipelineName()).getGitRepoSchema().getGitRepo();
        var getFileRequest = new GetFileRequest(gitRepoUrl, PIPELINE_FILE_NAME, gitCommitHash);
        try {
            return objectMapper.readValue(
                    objectMapper.writeValueAsString(
                            yamlObjectMapper.readValue(repoManagerApi.getFileFromRepo(getFileRequest, event.getOrgName(), event.getTeamName()), Object.class)
                    ),
                    PipelineData.class);
        } catch (JsonProcessingException e) {
            log.error("Could not parse YAML pipeline data file", e);
            throw new AtlasNonRetryableError(e);
        }
    }
}
