package com.greenops.pipelinereposerver.api.model.pipeline;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import com.greenops.pipelinereposerver.api.model.git.GitRepoSchema;

@JsonDeserialize(as = PipelineSchemaImpl.class)
public interface PipelineSchema {

    String getPipelineName();
    GitRepoSchema getGitRepoSchema();
    void setPipelineName(String name);
    void setGitRepoSchema(GitRepoSchema gitRepoSchema);
}
