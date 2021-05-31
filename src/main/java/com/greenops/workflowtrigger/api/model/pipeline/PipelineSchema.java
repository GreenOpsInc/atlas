package com.greenops.workflowtrigger.api.model.pipeline;

import com.greenops.workflowtrigger.api.model.git.GitRepoSchema;

public interface PipelineSchema {

    String getPipelineName();
    GitRepoSchema getGitRepoSchema();
    void setPipelineName(String name);
    void setGitRepoSchema(GitRepoSchema gitRepoSchema);
    //public
}
