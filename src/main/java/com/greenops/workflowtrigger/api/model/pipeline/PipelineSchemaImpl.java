package com.greenops.workflowtrigger.api.model.pipeline;

import com.greenops.workflowtrigger.api.model.git.GitRepoSchema;

public class PipelineSchemaImpl implements PipelineSchema {

    private String name;
    private GitRepoSchema gitRepoSchema;

    public PipelineSchemaImpl(String name, GitRepoSchema gitRepoSchema) {
        this.name = name;
        this.gitRepoSchema = gitRepoSchema;
    }

    @Override
    public String getPipelineName() {
        return name;
    }

    @Override
    public GitRepoSchema getGitRepoSchema() {
        return gitRepoSchema;
    }

    @Override
    public void setPipelineName(String name) {
        this.name = name;
    }

    @Override
    public void setGitRepoSchema(GitRepoSchema gitRepoSchema) {
        this.gitRepoSchema = gitRepoSchema;
    }
}
