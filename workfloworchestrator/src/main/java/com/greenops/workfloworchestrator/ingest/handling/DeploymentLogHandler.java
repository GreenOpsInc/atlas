package com.greenops.workfloworchestrator.ingest.handling;

import com.greenops.util.datamodel.auditlog.DeploymentLog;
import com.greenops.util.datamodel.event.Event;
import com.greenops.util.datamodel.clientmessages.ResourceGvk;
import com.greenops.util.datamodel.event.FailureEvent;

import java.util.List;

public interface DeploymentLogHandler {

    void updateStepDeploymentLog(Event event, String stepName, String argoApplicationName, String revisionHash);

    void initializeNewStepLog(Event event, String stepName, String pipelineUvn, String gitCommitVersion);

    void initializeNewRemediationLog(Event event, String stepName, String pipelineUvn, List<ResourceGvk> resourceGvkList);

    void markDeploymentSuccessful(Event event, String stepName);

    void markStepSuccessful(Event event, String stepName);

    void markStateRemediated(Event event, String stepName);

    void markStateRemediationFailed(Event event, String stepName);

    void markStepFailedWithFailedDeployment(Event event, String stepName);

    void markStepFailedWithBrokenTest(Event event, String stepName, String testName, String testLog);

    void markStepFailedWithProcessingError(FailureEvent event, String stepName, String error);

    boolean areParentStepsComplete(Event event, List<String> parentSteps);

    String getStepStatus(Event event);

    String makeRollbackDeploymentLog(Event event, String stepName, int rollbackLimit, boolean dryRun);

    String getCurrentGitCommitHash(Event event, String stepName);

    String getCurrentArgoRevisionHash(Event event, String stepName);

    String getCurrentPipelineUvn(Event event, String stepName);

    String getLastSuccessfulStepGitCommitHash(Event event, String stepName);

    String getLastSuccessfulDeploymentGitCommitHash(Event event, String stepName);

    DeploymentLog getLatestDeploymentLog(Event event, String stepName);
}
