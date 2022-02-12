package com.greenops.workfloworchestrator.ingest.handling;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.util.datamodel.auditlog.Log;
import com.greenops.util.datamodel.auditlog.PipelineInfo;
import com.greenops.util.datamodel.auditlog.RemediationLog;
import com.greenops.util.datamodel.clientmessages.ResourceGvk;
import com.greenops.util.datamodel.event.*;
import com.greenops.util.datamodel.git.GitRepoSchemaInfo;
import com.greenops.util.datamodel.pipeline.TeamSchema;
import com.greenops.util.datamodel.request.GetFileRequest;
import com.greenops.util.dbclient.DbClient;
import com.greenops.util.error.AtlasNonRetryableError;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.PipelineData;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.StepData;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.Test;
import com.greenops.workfloworchestrator.ingest.apiclient.reposerver.RepoManagerApi;
import com.greenops.workfloworchestrator.ingest.dbclient.DbKey;
import com.greenops.workfloworchestrator.ingest.kafka.KafkaClient;
import lombok.extern.slf4j.Slf4j;
import org.apache.logging.log4j.util.Strings;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

import static com.greenops.util.datamodel.event.ClientCompletionEvent.*;
import static com.greenops.workfloworchestrator.datamodel.pipelinedata.StepData.ROOT_STEP_NAME;
import static com.greenops.workfloworchestrator.datamodel.pipelinedata.StepData.createRootStep;
import static com.greenops.workfloworchestrator.ingest.handling.util.deployment.ArgoDeploymentInfo.NO_OP_ARGO_DEPLOYMENT;

@Slf4j
@Component
public class EventHandlerImpl implements EventHandler {

    static final String WATCH_ARGO_APPLICATION_KEY = "WatchArgoApplicationKey";
    public static final String WATCH_TEST_KEY = "WatchTestKey";
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
        log.info("Handling event of type {}", event.getClass().getName());
        var teamSchema = fetchTeamSchema(event);
        if (teamSchema == null) throw new AtlasNonRetryableError("The team doesn't exist");

        //This checks whether the event was from a previous pipeline run. If it is, it will be ignored.
        if (isStaleEvent(event)) {
            log.info("Event from pipeline {} is stale, ignoring...", event.getPipelineUvn());
            return;
        }
        if (event instanceof FailureEvent) {
            handleFailureEvent((FailureEvent) event);
            return;
        }

        var gitCommit = "";
        if (event instanceof TriggerStepEvent) {
            gitCommit = ((TriggerStepEvent) event).getGitCommitHash();
        } else if (event instanceof PipelineTriggerEvent) {
            gitCommit = ((PipelineTriggerEvent) event).getRevisionHash();
        } else if (!event.getStepName().equals(ROOT_STEP_NAME)) {
            gitCommit = deploymentLogHandler.getCurrentGitCommitHash(event, event.getStepName());
        }

        var tempGitRepoSchema = teamSchema.getPipelineSchema(event.getPipelineName()).getGitRepoSchema();
        var gitRepoSchemaInfo = new GitRepoSchemaInfo(tempGitRepoSchema.getGitRepo(), tempGitRepoSchema.getPathToRoot());
        var pipelineData = fetchPipelineData(event, gitRepoSchemaInfo, gitCommit);
        if (pipelineData == null) throw new AtlasNonRetryableError("The pipeline doesn't exist");

        if (event instanceof PipelineTriggerEvent) {
            handlePipelineTriggerEvent(pipelineData, gitRepoSchemaInfo, (PipelineTriggerEvent) event);
        } else if (event instanceof ClientCompletionEvent) {
            handleClientCompletionEvent(pipelineData, gitRepoSchemaInfo, (ClientCompletionEvent) event);
        } else if (event instanceof TestCompletionEvent) {
            handleTestCompletion(pipelineData, gitRepoSchemaInfo, (TestCompletionEvent) event);
        } else if (event instanceof ApplicationInfraTriggerEvent) {
            handleApplicationInfraTrigger(teamSchema, pipelineData, gitRepoSchemaInfo, (ApplicationInfraTriggerEvent) event);
        } else if (event instanceof ApplicationInfraCompletionEvent) {
            handleApplicationInfraCompletion(gitRepoSchemaInfo, pipelineData, (ApplicationInfraCompletionEvent) event);
        } else if (event instanceof TriggerStepEvent) {
            handleTriggerStep(pipelineData, gitRepoSchemaInfo, (TriggerStepEvent) event);
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

    private void handlePipelineTriggerEvent(PipelineData pipelineData, GitRepoSchemaInfo gitRepoSchemaInfo, PipelineTriggerEvent event) {
        var pipelineInfoKey = DbKey.makeDbPipelineInfoKey(event.getOrgName(), event.getTeamName(), event.getPipelineName());
        var pipelineInfo = dbClient.fetchLatestPipelineInfo(pipelineInfoKey);
        var listOfSteps = pipelineInfo != null ? pipelineInfo.getStepList() : null;
        PipelineData currentPipelineData;
        if (pipelineInfo != null && listOfSteps != null) {
            var latestUvn = pipelineInfo.getPipelineUvn();
            var progressing = false;
            var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), listOfSteps.get(0));
            var deploymentLog = dbClient.fetchLatestDeploymentLog(logKey);
            if (deploymentLog == null || !deploymentLog.getPipelineUniqueVersionNumber().equals(latestUvn)) {
                if (pipelineInfo.getErrors().size() == 0) {
                    //If there is a list but the deployment log is null it means the first step has not been triggered yet
                    queueNewPipelineRun(event, latestUvn);
                } else {
                    //Errors combined with no logs for the first step means there was an error and the pipeline is complete
                    initializeNewPipelineRun(event, gitRepoSchemaInfo, pipelineData);
                }
                return;
            }
            currentPipelineData = fetchPipelineData(event, gitRepoSchemaInfo, deploymentLog.getGitCommitVersion());

            //There are two options for starting steps:
            //1. Either the pipeline was run normally, and the starting steps are the root steps
            //2. A sub-pipeline was run, and there is an arbitrary starting step
            var orderedStepsBfs = new ArrayList<String>();
            var rootSteps = currentPipelineData.getChildrenSteps(ROOT_STEP_NAME);
            for (var stepName : rootSteps) {
                if (pipelineInfo.getStepList().contains(stepName)) {
                    orderedStepsBfs.add(stepName);
                }
            }
            if (orderedStepsBfs.size() == 0) {
                orderedStepsBfs.add(pipelineInfo.getStepList().get(0));
            }
            //Iterates through the logs to verify completion
            var idx = 0;
            while (idx < orderedStepsBfs.size()) {
                logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), orderedStepsBfs.get(idx));
                deploymentLog = dbClient.fetchLatestDeploymentLog(logKey);
                if (deploymentLog == null || !deploymentLog.getPipelineUniqueVersionNumber().equals(latestUvn)) {
                    progressing = true;
                    break;
                }
                if (deploymentLog.getPipelineUniqueVersionNumber().equals(latestUvn) && deploymentLog.getStatus().equals(Log.LogStatus.PROGRESSING.name())) {
                    progressing = true;
                    break;
                } else if (deploymentLog.getPipelineUniqueVersionNumber().equals(latestUvn) && deploymentLog.getStatus().equals(Log.LogStatus.FAILURE.name())) {
                    var stepData = currentPipelineData.getStep(orderedStepsBfs.get(idx));
                    //If the step has rollbacks enabled but has failed & not hit its rollback limit, the pipeline is still progressing
                    //The second part of the expression ensures that there is not a valid version to rollback to (if there isn't, the rollback limit may not have been hit)
                    if (stepData.getRollbackLimit() > deploymentLog.getUniqueVersionInstance()
                            && !deploymentLogHandler.makeRollbackDeploymentLog(event, stepData.getName(), stepData.getRollbackLimit(), true).isBlank()) {
                        progressing = true;
                        break;
                    }
                } else if (deploymentLog.getPipelineUniqueVersionNumber().equals(latestUvn) && deploymentLog.getStatus().equals(Log.LogStatus.SUCCESS.name())) {
                    orderedStepsBfs.addAll(
                            currentPipelineData.getChildrenSteps(orderedStepsBfs.get(idx)).stream().filter(
                                    stepName -> pipelineInfo.getStepList().contains(stepName)
                            ).collect(Collectors.toList())
                    );
                }
                idx++;
            }
            if (progressing) {
                queueNewPipelineRun(event, latestUvn);
                return;
            }
        }
        initializeNewPipelineRun(event, gitRepoSchemaInfo, pipelineData);
    }

    private void queueNewPipelineRun(PipelineTriggerEvent event, String latestUvn) {
        kafkaClient.sendMessage(event);
        log.info("Pipeline {} in progress, queueing up pipeline {}.", latestUvn, event.getPipelineUvn());
    }

    private void initializeNewPipelineRun(PipelineTriggerEvent event, GitRepoSchemaInfo gitRepoSchemaInfo, PipelineData pipelineData) {
        var pipelineInfoKey = DbKey.makeDbPipelineInfoKey(event.getOrgName(), event.getTeamName(), event.getPipelineName());
        log.info("Starting new run for pipeline {}", event.getPipelineName());
        if (event.getStepName().equals(ROOT_STEP_NAME)) {
            dbClient.insertValueInList(pipelineInfoKey, new PipelineInfo(event.getPipelineUvn(), List.of(), pipelineData.getAllStepsOrdered()));
            triggerNextSteps(pipelineData, createRootStep(), gitRepoSchemaInfo, event);
        } else {
            dbClient.insertValueInList(pipelineInfoKey, new PipelineInfo(event.getPipelineUvn(), List.of(), List.of(event.getStepName())));
            kafkaClient.sendMessage(new TriggerStepEvent(event.getOrgName(), event.getTeamName(), event.getPipelineName(), event.getStepName(), event.getPipelineUvn(), event.getRevisionHash(), false));
        }
    }

    private void handleClientCompletionEvent(PipelineData pipelineData, GitRepoSchemaInfo gitRepoSchemaInfo, ClientCompletionEvent event) {
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
                        if (step.getRollbackLimit() > 0) {
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
                deploymentHandler.triggerStateRemediation(event, gitRepoSchemaInfo, step, deploymentLog.getArgoApplicationName(), deploymentLog.getArgoRevisionHash(), resourceGvkList);
            } else if (event.getHealthStatus().equals(HEALTHY)) {
                deploymentLogHandler.markStateRemediated(event, step.getName());
            }
        } else {
            deploymentLogHandler.updateStepDeploymentLog(event, event.getStepName(), event.getArgoName(), event.getRevisionHash());
            if (event.getHealthStatus().equals(DEGRADED) || event.getHealthStatus().equals(UNKNOWN)) {
                deploymentLogHandler.markStepFailedWithFailedDeployment(event, event.getStepName());
                if (step.getRollbackLimit() > 0) rollback(event);
                return;
            }
            deploymentLogHandler.markDeploymentSuccessful(event, event.getStepName());

            var afterTestsExist = step.getTests().stream().anyMatch(test -> !test.shouldExecuteBefore());
            if (afterTestsExist) {
                testHandler.triggerTest(gitRepoSchemaInfo, step, false, deploymentLogHandler.getCurrentGitCommitHash(event, step.getName()), event);
            } else {
                triggerNextSteps(pipelineData, step, gitRepoSchemaInfo, event);
            }
        }
    }

    private void handleTestCompletion(PipelineData pipelineData, GitRepoSchemaInfo gitRepoSchemaInfo, TestCompletionEvent event) {
        var step = pipelineData.getStep(event.getStepName());
        if (!event.getSuccessful()) {
            deploymentLogHandler.markStepFailedWithBrokenTest(event, event.getStepName(), getTestNameFromNumber(step, event.getTestNumber()), event.getLog());
            if (step.getRollbackLimit() > 0) rollback(event);
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
        } else if (!completedTest.shouldExecuteBefore() && completedTestNumber == step.getTests().size() - 1) {
            //If there are before tasks, the numbering for "after" tasks won't be exact.
            //The completed test number may go past the number of after tests.
            triggerNextSteps(pipelineData, step, gitRepoSchemaInfo, event);
        } else if ((completedTest.shouldExecuteBefore() && completedTestNumber < tests.size())
                || (!completedTest.shouldExecuteBefore() && completedTestNumber < step.getTests().size())) {
            testHandler.createAndRunTest(
                    step.getClusterName(),
                    step,
                    gitRepoSchemaInfo,
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

    private void handleApplicationInfraTrigger(TeamSchema teamSchema, PipelineData pipelineData, GitRepoSchemaInfo gitRepoSchemaInfo, ApplicationInfraTriggerEvent event) {
        //Right now it is assumed that the step names do not change
        var stepData = pipelineData.getStep(event.getStepName());
        if (stepData.getOtherDeploymentsPath() == null) {
            kafkaClient.sendMessage(new ApplicationInfraCompletionEvent(event.getOrgName(), event.getTeamName(), event.getPipelineName(), event.getPipelineUvn(), event.getStepName(), true));
            return;
        }
        var oldGitCommitHash = stepData.getRollbackLimit() > 0
                ? deploymentLogHandler.getLastSuccessfulStepGitCommitHash(event, event.getStepName())
                //TODO: Add deploymentInfraSuccessful variable in deployment log, and replace method below with one that checks for the deployment infra deploying successfully
                : deploymentLogHandler.getLastSuccessfulDeploymentGitCommitHash(event, event.getStepName());
        if (oldGitCommitHash != null) {
            var oldPipelineData = fetchPipelineData(event, gitRepoSchemaInfo, oldGitCommitHash);
            var oldStepData = oldPipelineData.getStep(event.getStepName());
            //TODO: Two separate events cannot be sent for deleting then deploying. The deployment handler has to version application infrastructure with Kubernetes labels
            deploymentHandler.deleteApplicationInfrastructure(event, gitRepoSchemaInfo, oldStepData, oldGitCommitHash);
        }
        deploymentHandler.deployApplicationInfrastructure(event, gitRepoSchemaInfo, pipelineData.getStep(event.getStepName()), deploymentLogHandler.getCurrentGitCommitHash(event, event.getStepName()));
    }

    private void handleApplicationInfraCompletion(GitRepoSchemaInfo gitRepoSchemaInfo, PipelineData pipelineData, ApplicationInfraCompletionEvent event) {
        var stepData = pipelineData.getStep(event.getStepName());
        if (!event.isSuccess() && stepData.getRollbackLimit() > 0) rollback(event);
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepData.getName());
        var deploymentLog = dbClient.fetchLatestDeploymentLog(logKey);

        var argoDeploymentInfo = NO_OP_ARGO_DEPLOYMENT;
        if ((stepData.getArgoApplicationPath() != null || stepData.getArgoApplication() != null) && deploymentLog.getUniqueVersionInstance() > 0) {
            deploymentHandler.rollbackArgoApplication(event, gitRepoSchemaInfo, stepData, deploymentLog.getArgoApplicationName(), deploymentLog.getArgoRevisionHash());
            return;
        } else if (stepData.getArgoApplicationPath() != null || stepData.getArgoApplication() != null) {
            var argoRevisionHash = deploymentLogHandler.getCurrentArgoRevisionHash(event, stepData.getName());
            deploymentHandler.deployArgoApplication(event, gitRepoSchemaInfo, pipelineData, stepData.getName(), argoRevisionHash, deploymentLogHandler.getCurrentGitCommitHash(event, stepData.getName()));
        }
    }

    private void handleTriggerStep(PipelineData pipelineData, GitRepoSchemaInfo gitRepoSchemaInfo, TriggerStepEvent event) {
        var gitCommit = event.getGitCommitHash();

        //First cancel any pending remediation steps
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), event.getStepName());
        var latestLog = dbClient.fetchLatestLog(logKey);
        if (latestLog instanceof RemediationLog && latestLog.getStatus().equals(Log.LogStatus.PROGRESSING.name())) {
            latestLog.setStatus(Log.LogStatus.CANCELLED.name());
            dbClient.updateHeadInList(logKey, latestLog);
        }

        if (event.isRollback()) {
            log.info("Handling rollback trigger");
            gitCommit = deploymentLogHandler.makeRollbackDeploymentLog(event, event.getStepName(), pipelineData.getStep(event.getStepName()).getRollbackLimit(), false);
            if (gitCommit.isEmpty()) {
                log.info("Could not find stable deployment to rollback to");
                //Means there is no stable version that can be found.
                return;
            }
            pipelineData = fetchPipelineData(event, gitRepoSchemaInfo, gitCommit);
        } else {
            deploymentLogHandler.initializeNewStepLog(
                    event,
                    event.getStepName(),
                    event.getPipelineUvn(),
                    gitCommit
            );
        }

        var stepData = pipelineData.getStep(event.getStepName());

        var beforeTestsExist = stepData.getTests().stream().anyMatch(Test::shouldExecuteBefore);
        if (beforeTestsExist) {
            testHandler.triggerTest(gitRepoSchemaInfo, stepData, true, gitCommit, event);
            return;
        }

        if (stepData.getOtherDeploymentsPath() != null || stepData.getArgoApplicationPath() != null) {
            triggerAppInfraDeploy(stepData.getName(), event);
            return;
        }

        var afterTestsExist = stepData.getTests().stream().anyMatch(test -> !test.shouldExecuteBefore());
        if (afterTestsExist) {
            testHandler.triggerTest(gitRepoSchemaInfo, stepData, false, gitCommit, event);
            return;
        }
    }

    private void handleFailureEvent(FailureEvent event) {
        deploymentLogHandler.markStepFailedWithProcessingError(event, event.getStepName(), event.getError());
    }

    private void triggerNextSteps(PipelineData pipelineData, StepData step, GitRepoSchemaInfo gitRepoSchemaInfo, Event event) {
        var currentGitCommit = "";
        if (event instanceof PipelineTriggerEvent) {
            currentGitCommit = ((PipelineTriggerEvent) event).getRevisionHash();
        } else if (!step.getName().equals(ROOT_STEP_NAME)) {
            deploymentLogHandler.markStepSuccessful(event, event.getStepName());
            currentGitCommit = deploymentLogHandler.getCurrentGitCommitHash(event, step.getName());
        }

        if (isSubPipelineRun(event)) {
            return;
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

    private boolean isSubPipelineRun(Event event) {
        var pipelineInfo = dbClient.fetchLatestPipelineInfo(DbKey.makeDbPipelineInfoKey(event.getOrgName(), event.getTeamName(), event.getPipelineName()));
        if (pipelineInfo == null) {
            throw new AtlasNonRetryableError("Pipeline info does not exist when attempting to verify non-sub pipeline run");
        }
        //IS a sub-pipeline run, given that only one step is being run and the step names equal each other
        //This could also catch cases where there is a 1-step pipeline, but the behavior will be the same
        return pipelineInfo.getStepList().size() == 1 && event.getStepName().equals(pipelineInfo.getStepList().get(0));
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

    private PipelineData fetchPipelineData(Event event, GitRepoSchemaInfo gitRepoSchemaInfo, String gitCommitHash) {
        var getFileRequest = new GetFileRequest(gitRepoSchemaInfo, PIPELINE_FILE_NAME, gitCommitHash);
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

    private String getTestNameFromNumber(StepData stepData, Integer testNumber){
        return Strings.join(List.of(stepData.getName(), stepData.getTests().get(testNumber).getPath(), testNumber.toString()), '-');
    }
}
