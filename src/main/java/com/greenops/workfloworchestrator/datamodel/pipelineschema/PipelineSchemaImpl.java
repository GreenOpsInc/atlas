package com.greenops.workfloworchestrator.datamodel.pipelineschema;

import com.greenops.workfloworchestrator.datamodel.git.GitRepoSchema;

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
