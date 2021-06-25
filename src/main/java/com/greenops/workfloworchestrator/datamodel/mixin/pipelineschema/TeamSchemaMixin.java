package com.greenops.workfloworchestrator.datamodel.mixin.pipelineschema;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.greenops.workfloworchestrator.datamodel.pipelineschema.PipelineSchema;

import java.util.List;

public abstract class TeamSchemaMixin {

    @JsonProperty("teamName")
    String teamName;

    @JsonProperty("parentTeam")
    String parentTeam;

    @JsonProperty("orgName")
    String orgName;

    @JsonProperty("pipelines")
    List<PipelineSchema> pipelines;

    @JsonCreator
    private TeamSchemaMixin(@JsonProperty("teamName") String teamName,
                            @JsonProperty("parentTeam") String parentTeam,
                            @JsonProperty("orgName") String orgName,
                            @JsonProperty("pipelines") List<PipelineSchema> pipelines) {
    }

    @JsonIgnore
    abstract List<String> getPipelineNames();

    @JsonIgnore
    abstract List<String> getPipelineSchemas();

    @JsonIgnore
    abstract PipelineSchema getPipelineSchema(String pipelineName);
}
