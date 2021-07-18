package com.greenops.workfloworchestrator.ingest.handling;

import com.greenops.workfloworchestrator.datamodel.event.Event;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.StepData;
import com.greenops.workfloworchestrator.ingest.handling.util.deployment.ArgoDeploymentInfo;

public interface DeploymentHandler {

    boolean deployApplicationInfrastructure(Event event, String pipelineRepoUrl, StepData stepData);
    ArgoDeploymentInfo deployArgoApplication(Event event, String pipelineRepoUrl, StepData stepData);
    boolean rollbackArgoApplication(Event event, String pipelineRepoUrl, StepData stepData, String argoApplicationName, int argoRevisionId);
}
