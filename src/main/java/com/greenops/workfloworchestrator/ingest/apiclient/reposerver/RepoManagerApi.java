package com.greenops.workfloworchestrator.ingest.apiclient.reposerver;

import com.greenops.workfloworchestrator.datamodel.requests.GetFileRequest;

public interface RepoManagerApi {
    public String getFileFromRepo(GetFileRequest getFileRequest, String orgName, String teamName);
    public String getCurrentPipelineCommitHash(String gitRepoUrl, String orgName, String teamName);
    public void resetRepoVersion(String gitCommit, String gitRepoUrl, String orgName, String teamName);
}
