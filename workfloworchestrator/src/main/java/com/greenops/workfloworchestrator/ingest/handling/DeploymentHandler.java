package com.greenops.workfloworchestrator.ingest.handling;

import com.greenops.util.datamodel.clientmessages.ResourceGvk;
import com.greenops.util.datamodel.event.Event;
import com.greenops.util.datamodel.git.GitRepoSchemaInfo;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.PipelineData;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.StepData;

import java.util.List;

public interface DeploymentHandler {

    void deleteApplicationInfrastructure(Event event, GitRepoSchemaInfo gitRepoSchemaInfo, StepData stepData, String gitCommitHash);

    void deployApplicationInfrastructure(Event event, GitRepoSchemaInfo gitRepoSchemaInfo, StepData stepData, String gitCommitHash);

    void deployArgoApplication(Event event, GitRepoSchemaInfo gitRepoSchemaInfo, PipelineData pipelineData, String stepName, String argoRevisionHash, String gitCommitHash);

    void rollbackArgoApplication(Event event, GitRepoSchemaInfo gitRepoSchemaInfo, StepData stepData, String argoApplicationName, String argoRevisionHash);

    void triggerStateRemediation(Event event, GitRepoSchemaInfo gitRepoSchemaInfo, StepData stepData, String argoApplicationName, String argoRevisionHash, List<ResourceGvk> resourceStatuses);

    boolean rollbackInPipelineExists(Event event, PipelineData pipelineData, String stepName);
}
