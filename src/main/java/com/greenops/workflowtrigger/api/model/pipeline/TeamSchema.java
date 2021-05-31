package com.greenops.workflowtrigger.api.model.pipeline;

import com.greenops.workflowtrigger.api.model.git.GitRepoSchema;

import java.util.List;

public interface TeamSchema {

    final static String ROOT_TEAM = "Root";

    String getTeamName();
    String getParentTeam();
    List<String> getPipelineNames();
    void setTeamName(String teamName);
    void setParentTeam(String parentTeam);
    void addPipeline(String pipelineName, GitRepoSchema gitRepoSchema);
    void addPipeline(PipelineSchema pipelineSchema);
    void removePipeline(String pipelineName);
    //TODO: get root parent team
}
