package com.greenops.pipelinereposerver.api.model.pipeline;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import com.greenops.pipelinereposerver.api.model.git.GitRepoSchema;

import java.util.List;

@JsonDeserialize(as = TeamSchemaImpl.class)
public interface TeamSchema {

    final static String ROOT_TEAM = "Root";

    String getTeamName();
    String getParentTeam();
    String getOrgName();
    List<String> getPipelineNames();
    List<PipelineSchema> getPipelineSchemas();
    void setTeamName(String teamName);
    void setParentTeam(String parentTeam);
    void addPipeline(String pipelineName, GitRepoSchema gitRepoSchema);
    void addPipeline(PipelineSchema pipelineSchema);
    void removePipeline(String pipelineName);
}
