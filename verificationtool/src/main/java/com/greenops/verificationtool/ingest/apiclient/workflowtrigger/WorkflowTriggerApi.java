package com.greenops.verificationtool.ingest.apiclient.workflowtrigger;

public interface WorkflowTriggerApi {
    void createTeam(String orgName, String parentTeamName, String teamName);

    void createPipeline(String orgName, String pipelineName, String teamName, String gitRepoUrl, String pathToRoot);

    void syncPipeline(String orgName, String pipelineName, String teamName, String gitRepoUrl, String pathToRoot);

    String getPipelineStatus(String orgName, String pipelineName, String teamName);

    String getStepLevelStatus(String orgName, String pipelineName, String teamName, String stepName, Integer count);
}
