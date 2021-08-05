package com.greenops.workfloworchestrator.ingest.handling;

import com.greenops.util.datamodel.event.Event;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.StepData;
import com.greenops.workfloworchestrator.ingest.handling.util.deployment.ArgoDeploymentInfo;

public interface DeploymentHandler {

    void deleteApplicationInfrastructure(Event event, String pipelineRepoUrl, StepData stepData, String gitCommitHash);
    void deployApplicationInfrastructure(Event event, String pipelineRepoUrl, StepData stepData, String gitCommitHash);
    ArgoDeploymentInfo deployArgoApplication(Event event, String pipelineRepoUrl, StepData stepData, String gitCommitHash);
    void rollbackArgoApplication(Event event, String pipelineRepoUrl, StepData stepData, String argoApplicationName, String argoRevisionHash);
}
