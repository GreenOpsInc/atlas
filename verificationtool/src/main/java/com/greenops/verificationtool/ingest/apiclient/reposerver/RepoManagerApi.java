package com.greenops.verificationtool.ingest.apiclient.reposerver;

import com.greenops.util.datamodel.request.GetFileRequest;

public interface RepoManagerApi {
    public static final String ROOT_COMMIT = "ROOT_COMMIT";

    public String getFileFromRepo(GetFileRequest getFileRequest, String orgName, String teamName);
    public String getCurrentPipelineCommitHash(String gitRepoUrl, String orgName, String teamName);
    public void resetRepoVersion(String gitCommit, String gitRepoUrl, String orgName, String teamName);
}