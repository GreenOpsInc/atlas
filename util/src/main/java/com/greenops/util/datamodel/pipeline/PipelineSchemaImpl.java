package com.greenops.util.datamodel.pipeline;


import com.greenops.util.datamodel.git.GitRepoSchema;

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

    @Override
    public boolean equals(Object o) {
        if(o instanceof PipelineSchemaImpl){
            return ((PipelineSchemaImpl) o).gitRepoSchema.equals(gitRepoSchema)
                    && ((PipelineSchemaImpl) o).pipelineName.equals(pipelineName);
        }
        return false;
    }
}
