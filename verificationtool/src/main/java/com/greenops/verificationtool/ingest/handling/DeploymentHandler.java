package com.greenops.verificationtool.ingest.handling;

import com.greenops.util.datamodel.event.Event;
import com.greenops.verificationtool.datamodel.pipelinedata.PipelineData;
import com.greenops.verificationtool.datamodel.pipelinedata.StepData;
import com.greenops.util.datamodel.clientmessages.ResourceGvk;
import com.greenops.verificationtool.ingest.handling.util.deployment.ArgoDeploymentInfo;

import java.util.List;

public interface DeploymentHandler {

    void deleteApplicationInfrastructure(Event event, String pipelineRepoUrl, StepData stepData, String gitCommitHash);

    void deployApplicationInfrastructure(Event event, String pipelineRepoUrl, StepData stepData, String gitCommitHash);

    void deployArgoApplication(Event event, String pipelineRepoUrl, PipelineData pipelineData, String stepName, String argoRevisionHash, String gitCommitHash);

    void rollbackArgoApplication(Event event, String pipelineRepoUrl, StepData stepData, String argoApplicationName, String argoRevisionHash);

    void triggerStateRemediation(Event event, String pipelineRepoUrl, StepData stepData, String argoApplicationName, String argoRevisionHash, List<ResourceGvk> resourceStatuses);

    boolean rollbackInPipelineExists(Event event, PipelineData pipelineData, String stepName);
}
