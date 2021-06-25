package com.greenops.workfloworchestrator.datamodel.requests;

public class GetFileRequest {
    private final String gitRepoUrl;
    private final String filename;

    GetFileRequest(String gitRepoUrl, String filename) {
        this.gitRepoUrl = gitRepoUrl;
        this.filename = filename;
    }

    public String getGitRepoUrl() {
        return gitRepoUrl;
    }

    public String getFilename() {
        return filename;
    }
}
