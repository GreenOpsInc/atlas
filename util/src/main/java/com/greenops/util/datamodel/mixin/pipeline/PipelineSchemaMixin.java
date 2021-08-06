package com.greenops.util.datamodel.mixin.pipeline;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.greenops.util.datamodel.git.GitRepoSchema;

public abstract class PipelineSchemaMixin {

    @JsonProperty("name")
    private String name;

    @JsonProperty("gitRepoSchema")
    private GitRepoSchema gitRepoSchema;

    public PipelineSchemaMixin(@JsonProperty("pipelineName") String pipelineName, @JsonProperty("gitRepoSchema") GitRepoSchema gitRepoSchema) {}
}
