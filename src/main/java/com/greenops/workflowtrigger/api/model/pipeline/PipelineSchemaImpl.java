package com.greenops.workflowtrigger.api.model.pipeline;

import com.greenops.workflowtrigger.api.model.git.GitRepoSchema;

public class PipelineSchemaImpl implements PipelineSchema {

    private String pipelineName;
    private GitRepoSchema gitRepoSchema;

    public PipelineSchemaImpl(String pipelineName, GitRepoSchema gitRepoSchema) {
        this.pipelineName = pipelineName;
        this.gitRepoSchema = gitRepoSchema;
    }

    @Override
    public String getPipelineName() {
        return pipelineName;
    }

    @Override
    public GitRepoSchema getGitRepoSchema() {
        return gitRepoSchema;
    }

    @Override
    public void setPipelineName(String pipelineName) {
        this.pipelineName = pipelineName;
    }

    @Override
    public void setGitRepoSchema(GitRepoSchema gitRepoSchema) {
        this.gitRepoSchema = gitRepoSchema;
    }
}
