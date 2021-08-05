package com.greenops.util.datamodel.pipeline;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import com.greenops.util.datamodel.git.GitRepoSchema;

@JsonDeserialize(as = PipelineSchemaImpl.class)
public interface PipelineSchema {

    String getPipelineName();
    GitRepoSchema getGitRepoSchema();
    void setPipelineName(String name);
    void setGitRepoSchema(GitRepoSchema gitRepoSchema);
}
