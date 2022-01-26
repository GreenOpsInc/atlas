package com.greenops.workfloworchestrator.ingest.apiclient.reposerver;

import com.greenops.util.datamodel.git.GitRepoSchemaInfo;
import com.greenops.util.datamodel.request.GetFileRequest;

public interface RepoManagerApi {

    public String getFileFromRepo(GetFileRequest getFileRequest, String orgName, String teamName);
    public void resetRepoVersion(String gitCommit, GitRepoSchemaInfo gitRepoSchemaInfo, String orgName, String teamName);
}
