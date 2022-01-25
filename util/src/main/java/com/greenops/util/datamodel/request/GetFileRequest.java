package com.greenops.util.datamodel.request;

import com.greenops.util.datamodel.git.GitRepoSchemaInfo;

public class GetFileRequest {
    private final GitRepoSchemaInfo gitRepoSchemaInfo;
    private final String filename;
    private final String gitCommitHash;

    public GetFileRequest(GitRepoSchemaInfo gitRepoSchemaInfo, String filename, String gitCommitHash) {
        this.gitRepoSchemaInfo = gitRepoSchemaInfo;
        this.filename = filename;
        this.gitCommitHash = gitCommitHash;
    }

    public GitRepoSchemaInfo getGitRepoSchemaInfo() {
        return gitRepoSchemaInfo;
    }

    public String getFilename() {
        return filename;
    }

    public String getGitCommitHash() {
        return gitCommitHash;
    }
}
