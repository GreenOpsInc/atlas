package com.greenops.workfloworchestrator.ingest.apiclient.reposerver;

import com.greenops.workfloworchestrator.datamodel.requests.GetFileRequest;

public interface RepoManagerApi {
    public String getFileFromRepo(GetFileRequest getFileRequest, String orgName, String teamName);
}
