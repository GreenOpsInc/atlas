package com.greenops.workflowtrigger.api.model.pipeline;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import com.greenops.workflowtrigger.api.model.git.GitRepoSchema;

import java.util.List;

@JsonDeserialize(as = TeamSchemaImpl.class)
public interface TeamSchema {

    final static String ORG_NAME = "Temporary"; //TODO: This should be updated once we know how users are going to specify the org name
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
