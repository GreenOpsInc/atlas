package com.greenops.workflowtrigger.api.model.mixin.pipeline;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.greenops.workflowtrigger.api.model.git.GitRepoSchema;

public abstract class PipelineSchemaMixin {

    @JsonProperty("name")
    private String name;

    @JsonProperty("gitRepoSchema")
    private GitRepoSchema gitRepoSchema;

    public PipelineSchemaMixin(@JsonProperty("pipelineName") String pipelineName, @JsonProperty("gitRepoSchema") GitRepoSchema gitRepoSchema) {}
}
