package com.greenops.workfloworchestrator.datamodel.mixin.pipelineschema;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.greenops.workfloworchestrator.datamodel.git.GitRepoSchema;

public abstract class PipelineSchemaMixin {

    @JsonProperty("name")
    private String name;

    @JsonProperty("gitRepoSchema")
    private GitRepoSchema gitRepoSchema;

    public PipelineSchemaMixin(@JsonProperty("pipelineName") String pipelineName, @JsonProperty("gitRepoSchema") GitRepoSchema gitRepoSchema) {}
}
