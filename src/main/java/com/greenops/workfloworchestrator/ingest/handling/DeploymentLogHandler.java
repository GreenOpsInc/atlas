package com.greenops.workfloworchestrator.ingest.handling;

import com.greenops.workfloworchestrator.datamodel.event.Event;

public interface DeploymentLogHandler {

    void updateStepDeploymentLog(Event event, String stepName, String argoApplicationName, int revisionId);

    void initializeNewStepLog(Event event, String stepName, String gitCommitVersion);

    void markDeploymentSuccessful(Event event, String stepName);

    void markStepSuccessful(Event event, String stepName);

    void markStepFailedWithBrokenTest(Event event, String stepName, String testName, String testLog);

    boolean areParentStepsComplete(String stepName);

    String makeRollbackDeploymentLog(Event event, String stepName);
}
