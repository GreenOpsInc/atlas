package com.greenops.workfloworchestrator.ingest.handling;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.util.datamodel.auditlog.Log;
import com.greenops.util.datamodel.auditlog.RemediationLog;
import com.greenops.util.datamodel.event.*;
import com.greenops.util.datamodel.pipeline.TeamSchema;
import com.greenops.util.datamodel.request.GetFileRequest;
import com.greenops.util.dbclient.DbClient;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.PipelineData;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.StepData;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.Test;
import com.greenops.util.datamodel.clientmessages.ResourceGvk;
import com.greenops.workfloworchestrator.error.AtlasNonRetryableError;
import com.greenops.workfloworchestrator.ingest.apiclient.reposerver.RepoManagerApi;
import com.greenops.workfloworchestrator.ingest.dbclient.DbKey;
import com.greenops.workfloworchestrator.ingest.kafka.KafkaClient;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.stream.Collectors;

import static com.greenops.util.datamodel.event.ClientCompletionEvent.*;
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
        var gitCommit = ROOT_COMMIT;
        if (event instanceof TriggerStepEvent) {
            gitCommit = ((TriggerStepEvent) event).getGitCommitHash();
        } else if (!(event.getStepName().equals(ROOT_STEP_NAME) || event instanceof PipelineTriggerEvent)) {
            gitCommit = deploymentLogHandler.getCurrentGitCommitHash(event, event.getStepName());
        }
        var gitRepoUrl = teamSchema.getPipelineSchema(event.getPipelineName()).getGitRepoSchema().getGitRepo();
        var pipelineData = fetchPipelineData(event, gitRepoUrl, gitCommit);
        if (pipelineData == null) throw new AtlasNonRetryableError("The pipeline doesn't exist");

        //This checks whether the event was from a previous pipeline run. If it is, it will be ignored.
        if (isStaleEvent(event)) {
            log.info("Event from pipeline {} is stale, ignoring...", event.getPipelineUvn());
            return;
        }
        if (event instanceof PipelineTriggerEvent) {
            log.info("Handling event of type {}", PipelineTriggerEvent.class.getName());
            handlePipelineTriggerEvent(pipelineData, gitRepoUrl, (PipelineTriggerEvent) event);
        } else if (event instanceof ClientCompletionEvent) {
            log.info("Handling event of type {}", ClientCompletionEvent.class.getName());
            handleClientCompletionEvent(pipelineData, gitRepoUrl, (ClientCompletionEvent) event);
        } else if (event instanceof TestCompletionEvent) {
            log.info("Handling event of type {}", TestCompletionEvent.class.getName());
            handleTestCompletion(pipelineData, gitRepoUrl, (TestCompletionEvent) event);
        } else if (event instanceof ApplicationInfraTriggerEvent) {
            log.info("Handling event of type {}", ApplicationInfraTriggerEvent.class.getName());
            handleApplicationInfraTrigger(teamSchema, pipelineData, gitRepoUrl, (ApplicationInfraTriggerEvent) event);
        } else if (event instanceof ApplicationInfraCompletionEvent) {
            log.info("Handling event of type {}", ApplicationInfraCompletionEvent.class.getName());
            handleApplicationInfraCompletion(gitRepoUrl, pipelineData, (ApplicationInfraCompletionEvent) event);
        } else if (event instanceof TriggerStepEvent) {
            log.info("Handling event of type {}", TriggerStepEvent.class.getName());
            handleTriggerStep(pipelineData, gitRepoUrl, (TriggerStepEvent) event);
        } else if (event instanceof FailureEvent) {
            log.info("Handling event of type {}", FailureEvent.class.getName());
            handleFailureEvent(pipelineData, gitRepoUrl, (FailureEvent) event);
        }
    }

    private boolean isStaleEvent(Event event) {
        if (event instanceof PipelineTriggerEvent) return false;
        var deploymentLog = deploymentLogHandler.getLatestDeploymentLog(event, event.getStepName());
        if (deploymentLog != null
                && (deploymentLog.getStatus().equals(Log.LogStatus.SUCCESS.name()) || deploymentLog.getStatus().equals(Log.LogStatus.FAILURE.name()) || deploymentLog.getStatus().equals(Log.LogStatus.CANCELLED.name()))
                && (event instanceof TestCompletionEvent || event instanceof ApplicationInfraTriggerEvent || event instanceof ApplicationInfraCompletionEvent || event instanceof FailureEvent)) {
            return true;
        } else if (deploymentLog != null && deploymentLog.getStatus().equals(Log.LogStatus.PROGRESSING.name()) && event instanceof TriggerStepEvent) {
            return true;
        }
        var currentUvn = deploymentLogHandler.getCurrentPipelineUvn(event, event.getStepName());
        if (!event.getPipelineUvn().equals(currentUvn) && event instanceof TriggerStepEvent) {
            return false;
        }
        return currentUvn != null && !event.getPipelineUvn().equals(currentUvn);
    }

    private void handlePipelineTriggerEvent(PipelineData pipelineData, String pipelineRepoUrl, PipelineTriggerEvent event) {
        var listOfSteps = dbClient.fetchStringList(DbKey.makeDbListOfStepsKey(event.getOrgName(), event.getTeamName(), event.getPipelineName()));
        if (listOfSteps != null) {
            var latestUvn = "";
            var hasNonZeroUvi = false;
            var incomplete = false;
            var progressing = false;
            for (var step : listOfSteps) {
                var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), step);
                var deploymentLog = dbClient.fetchLatestDeploymentLog(logKey);
                if (deploymentLog == null) {
                    incomplete = true;
                    continue;
                }
                if (latestUvn.isEmpty()) {
                    latestUvn = deploymentLog.getPipelineUniqueVersionNumber();
                }
                if (deploymentLog.getPipelineUniqueVersionNumber().equals(latestUvn) && deploymentLog.getStatus().equals(Log.LogStatus.PROGRESSING.name())) {
                    progressing = true;
                    break;
                } else if (deploymentLog.getPipelineUniqueVersionNumber().equals(latestUvn) && deploymentLog.getUniqueVersionInstance() > 0) {
                    hasNonZeroUvi = true;
                } else if (!deploymentLog.getPipelineUniqueVersionNumber().equals(latestUvn)) {
                    incomplete = true;
                }
            }
            if ((!hasNonZeroUvi && incomplete) || progressing) {
                kafkaClient.sendMessage(event);
                log.info("Pipeline {} in progress, queueing up pipeline {}.", latestUvn, event.getPipelineUvn());
                return;
            }
        }
        dbClient.storeValue(DbKey.makeDbListOfStepsKey(event.getOrgName(), event.getTeamName(), event.getPipelineName()), pipelineData.getAllStepsOrdered());
        log.info("Starting new run for pipeline {}", event.getPipelineName());
        triggerNextSteps(pipelineData, createRootStep(), pipelineRepoUrl, event);
    }

    private void handleClientCompletionEvent(PipelineData pipelineData, String pipelineRepoUrl, ClientCompletionEvent event) {
        if (event.getHealthStatus().equals(PROGRESSING)) {
            return;
        }

        var step = pipelineData.getStep(event.getStepName());
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), event.getStepName());
        var latestLog = dbClient.fetchLatestLog(logKey);
        if (latestLog != null && latestLog.getStatus().equals(Log.LogStatus.CANCELLED.name())) return;
        var deploymentLog = dbClient.fetchLatestDeploymentLog(logKey);
        if (deploymentLog == null) return;
        if (deploymentLog.getStatus().equals(Log.LogStatus.FAILURE.name()) || deploymentLog.getStatus().equals(Log.LogStatus.CANCELLED.name()))
            return;
        else if (deploymentLog.getStatus().equals(Log.LogStatus.SUCCESS.name())) {
            //If the deployment was successful (rollback or otherwise), these client completion events are for state remediation
            if (step.getRemediationLimit() == 0) {
                return;
            }

            if (event.getSyncStatus().equals(OUT_OF_SYNC) && event.getHealthStatus().equals(MISSING)) {
                if (deploymentHandler.rollbackInPipelineExists(event, pipelineData, step.getName())) {
                    return;
                }
            }

            var remediationLog = dbClient.fetchLatestRemediationLog(logKey);
            if (remediationLog != null && remediationLog.getStatus().equals(Log.LogStatus.CANCELLED.name())) return;

            //TODO: Right now, remediation is only based on health. We should be adding in pruning based on OutOfSync statuses as well.
            //TODO: Right now the events are being sent but are just being ignored.
            if (event.getHealthStatus().equals(DEGRADED) || event.getHealthStatus().equals(UNKNOWN)) {
//                    || event.getHealthStatus().equals(MISSING)) {
                if (remediationLog != null) {
                    if (remediationLog.getStatus().equals(Log.LogStatus.PROGRESSING.name())) {
                        deploymentLogHandler.markStateRemediationFailed(event, step.getName());
                    }
                    if (remediationLog.getUniqueVersionInstance() == step.getRemediationLimit()) {
                        //Reached remediation limit
                        if (step.getRollback()) {
                            rollback(event);
                        }
                        return;
                    }
                }
                var resourceGvkList = new ArrayList<ResourceGvk>();
                for (var resource : event.getResourceStatuses()) {
                    resourceGvkList.add(new ResourceGvk(resource.getResourceName(), resource.getResourceNamespace(), resource.getGroup(), resource.getVersion(), resource.getKind()));
                }
                deploymentLogHandler.initializeNewRemediationLog(event, step.getName(), deploymentLog.getPipelineUniqueVersionNumber(), resourceGvkList);
                deploymentHandler.triggerStateRemediation(event, pipelineRepoUrl, step, deploymentLog.getArgoApplicationName(), deploymentLog.getArgoRevisionHash(), resourceGvkList);
            } else if (event.getHealthStatus().equals(HEALTHY)) {
                deploymentLogHandler.markStateRemediated(event, step.getName());
            }
        } else {
            deploymentLogHandler.updateStepDeploymentLog(event, event.getStepName(), event.getArgoName(), event.getRevisionHash());
            if (event.getHealthStatus().equals(DEGRADED) || event.getHealthStatus().equals(UNKNOWN)) {
                deploymentLogHandler.markStepFailedWithFailedDeployment(event, event.getStepName());
                if (step.getRollback()) rollback(event);
                return;
            }
            deploymentLogHandler.markDeploymentSuccessful(event, event.getStepName());

            var afterTestsExist = step.getTests().stream().anyMatch(test -> !test.shouldExecuteBefore());
            if (afterTestsExist) {
                testHandler.triggerTest(pipelineRepoUrl, step, false, deploymentLogHandler.getCurrentGitCommitHash(event, step.getName()), event);
            } else {
                triggerNextSteps(pipelineData, step, pipelineRepoUrl, event);
            }
        }
    }

    private void handleTestCompletion(PipelineData pipelineData, String pipelineRepoUrl, TestCompletionEvent event) {
        var step = pipelineData.getStep(event.getStepName());
        if (!event.getSuccessful()) {
            deploymentLogHandler.markStepFailedWithBrokenTest(event, event.getStepName(), event.getTestName(), event.getLog());
            if (step.getRollback()) rollback(event);
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
                    step.getClusterName(),
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
        if (stepData.getOtherDeploymentsPath() == null) {
            kafkaClient.sendMessage(new ApplicationInfraCompletionEvent(event.getOrgName(), event.getTeamName(), event.getPipelineName(), event.getPipelineUvn(), event.getStepName(), true));
            return;
        }
        var oldGitCommitHash = stepData.getRollback()
                ? deploymentLogHandler.getLastSuccessfulStepGitCommitHash(event, event.getStepName())
                //TODO: Add deploymentInfraSuccessful variable in deployment log, and replace method below with one that checks for the deployment infra deploying successfully
                : deploymentLogHandler.getLastSuccessfulDeploymentGitCommitHash(event, event.getStepName());
        if (oldGitCommitHash != null) {
            var oldPipelineData = fetchPipelineData(event, pipelineRepoUrl, oldGitCommitHash);
            var oldStepData = oldPipelineData.getStep(event.getStepName());
            deploymentHandler.deleteApplicationInfrastructure(event, pipelineRepoUrl, oldStepData, oldGitCommitHash);
        }
        deploymentHandler.deployApplicationInfrastructure(event, pipelineRepoUrl, pipelineData.getStep(event.getStepName()), deploymentLogHandler.getCurrentGitCommitHash(event, event.getStepName()));
    }

    private void handleApplicationInfraCompletion(String pipelineRepoUrl, PipelineData pipelineData, ApplicationInfraCompletionEvent event) {
        var stepData = pipelineData.getStep(event.getStepName());
        if (!event.isSuccess() && stepData.getRollback()) rollback(event);
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepData.getName());
        var deploymentLog = dbClient.fetchLatestDeploymentLog(logKey);

        var argoDeploymentInfo = NO_OP_ARGO_DEPLOYMENT;
        if ((stepData.getArgoApplicationPath() != null || stepData.getArgoApplication() != null) && deploymentLog.getUniqueVersionInstance() > 0) {
            deploymentHandler.rollbackArgoApplication(event, pipelineRepoUrl, stepData, deploymentLog.getArgoApplicationName(), deploymentLog.getArgoRevisionHash());
            return;
        } else if (stepData.getArgoApplicationPath() != null || stepData.getArgoApplication() != null) {
            var argoRevisionHash = deploymentLogHandler.getCurrentArgoRevisionHash(event, stepData.getName());
            deploymentHandler.deployArgoApplication(event, pipelineRepoUrl, pipelineData, stepData.getName(), argoRevisionHash, deploymentLogHandler.getCurrentGitCommitHash(event, stepData.getName()));
        }
    }

    private void handleTriggerStep(PipelineData pipelineData, String pipelineRepoUrl, TriggerStepEvent event) {
        var gitCommit = ROOT_COMMIT;

        //First cancel any pending remediation steps
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), event.getStepName());
        var latestLog = dbClient.fetchLatestLog(logKey);
        if (latestLog instanceof RemediationLog && latestLog.getStatus().equals(Log.LogStatus.PROGRESSING.name())) {
            latestLog.setStatus(Log.LogStatus.CANCELLED.name());
            dbClient.updateHeadInList(logKey, latestLog);
        }

        if (event.isRollback()) {
            log.info("Handling rollback trigger");
            gitCommit = deploymentLogHandler.makeRollbackDeploymentLog(event, event.getStepName());
            if (gitCommit.isEmpty()) {
                log.info("Could not find stable deployment to rollback to");
                //Means there is no stable version that can be found.
                return;
            }
            pipelineData = fetchPipelineData(event, pipelineRepoUrl, gitCommit);
        } else {
            deploymentLogHandler.initializeNewStepLog(
                    event,
                    event.getStepName(),
                    event.getPipelineUvn(),
                    repoManagerApi.getCurrentPipelineCommitHash(pipelineRepoUrl, event.getOrgName(), event.getTeamName())
            );
        }

        var stepData = pipelineData.getStep(event.getStepName());

        var beforeTestsExist = stepData.getTests().stream().anyMatch(Test::shouldExecuteBefore);
        if (beforeTestsExist) {
            testHandler.triggerTest(pipelineRepoUrl, stepData, true, gitCommit, event);
            return;
        }

        if (stepData.getOtherDeploymentsPath() != null || stepData.getArgoApplicationPath() != null) {
            triggerAppInfraDeploy(stepData.getName(), event);
            return;
        }

        var afterTestsExist = stepData.getTests().stream().anyMatch(test -> !test.shouldExecuteBefore());
        if (afterTestsExist) {
            testHandler.triggerTest(pipelineRepoUrl, stepData, false, gitCommit, event);
            return;
        }
    }

    private void handleFailureEvent(PipelineData pipelineData, String pipelineRepoUrl, FailureEvent event) {
        deploymentLogHandler.markStepFailedWithProcessingError(event, event.getStepName(), event.getError());
        throw new AtlasNonRetryableError("Received failure event from client wrapper for step " + event.getStepName());
    }

    private void triggerNextSteps(PipelineData pipelineData, StepData step, String pipelineRepoUrl, Event event) {
        var currentGitCommit = ROOT_COMMIT;
        if (!step.getName().equals(ROOT_STEP_NAME)) {
            deploymentLogHandler.markStepSuccessful(event, event.getStepName());
            currentGitCommit = deploymentLogHandler.getCurrentGitCommitHash(event, step.getName());
        }
        var currentPipelineUvn = event.getPipelineUvn();

        var childrenSteps = pipelineData.getChildrenSteps(step.getName());
        var triggerStepEvents = new ArrayList<Event>();
        for (var stepName : childrenSteps) {
            var nextStep = pipelineData.getStep(stepName);
            var parentSteps = pipelineData.getParentSteps(stepName);
            if (deploymentLogHandler.areParentStepsComplete(event, parentSteps)) {
                triggerStepEvents.add(
                        new TriggerStepEvent(event.getOrgName(), event.getTeamName(), event.getPipelineName(), nextStep.getName(), currentPipelineUvn, currentGitCommit, false)
                );
            }
        }
        kafkaClient.sendMessage(triggerStepEvents);
    }

    private void triggerAppInfraDeploy(String stepName, Event event) {
        kafkaClient.sendMessage(new ApplicationInfraTriggerEvent(event.getOrgName(), event.getTeamName(), event.getPipelineName(), event.getPipelineUvn(), stepName));
    }

    private void rollback(Event event) {
        kafkaClient.sendMessage(
                new TriggerStepEvent(
                        event.getOrgName(),
                        event.getTeamName(),
                        event.getPipelineName(),
                        event.getStepName(),
                        event.getPipelineUvn(),
                        deploymentLogHandler.getCurrentGitCommitHash(event, event.getStepName()),
                        true
                )
        );
    }

    private TeamSchema fetchTeamSchema(Event event) {
        return dbClient.fetchTeamSchema(DbKey.makeDbTeamKey(event.getOrgName(), event.getTeamName()));
    }

    private PipelineData fetchPipelineData(Event event, String gitRepoUrl, String gitCommitHash) {
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
