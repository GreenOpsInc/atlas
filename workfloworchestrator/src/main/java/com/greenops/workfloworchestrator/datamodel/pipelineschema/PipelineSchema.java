package com.greenops.workfloworchestrator.datamodel.pipelineschema;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import com.greenops.workfloworchestrator.datamodel.git.GitRepoSchema;

@JsonDeserialize(as = PipelineSchemaImpl.class)
public interface PipelineSchema {

    String getPipelineName();
    GitRepoSchema getGitRepoSchema();
    void setPipelineName(String name);
    void setGitRepoSchema(GitRepoSchema gitRepoSchema);
}
