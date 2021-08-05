package com.greenops.pipelinereposerver.api.model.request;

public class GetFileRequest {
    private final String gitRepoUrl;
    private final String filename;
    private final String gitCommitHash;

    GetFileRequest(String gitRepoUrl, String filename, String gitCommitHash) {
        this.gitRepoUrl = gitRepoUrl;
        this.filename = filename;
        this.gitCommitHash = gitCommitHash;
    }

    public String getGitRepoUrl() {
        return gitRepoUrl;
    }

    public String getFilename() {
        return filename;
    }

    public String getGitCommitHash() {
        return gitCommitHash;
    }
}
