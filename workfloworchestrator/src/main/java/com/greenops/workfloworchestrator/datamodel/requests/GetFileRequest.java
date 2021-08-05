package com.greenops.workfloworchestrator.datamodel.requests;

public class GetFileRequest {

    private final String gitRepoUrl;
    private final String filename;
    private final String gitCommitHash;

    public GetFileRequest(String gitRepoUrl, String filename, String gitCommitHash) {
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
