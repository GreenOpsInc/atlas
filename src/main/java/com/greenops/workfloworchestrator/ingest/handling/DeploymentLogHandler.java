package com.greenops.workfloworchestrator.ingest.handling;

import com.greenops.workfloworchestrator.datamodel.event.Event;

public interface DeploymentLogHandler {

    boolean updateStepDeploymentLog(Event event, String stepName, String argoApplicationName, int revisionId);

    boolean initializeNewStepLog(Event event, String stepName, String gitCommitVersion);

    boolean markDeploymentSuccessful(Event event, String stepName);

    boolean markStepSuccessful(Event event, String stepName);

    boolean markStepFailedWithBrokenTest(Event event, String stepName, String testName, String testLog);

    boolean areParentStepsComplete(String stepName);

    String makeRollbackDeploymentLog(Event event, String stepName);
}
