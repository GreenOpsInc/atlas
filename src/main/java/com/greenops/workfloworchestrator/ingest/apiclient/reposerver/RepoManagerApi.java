package com.greenops.workfloworchestrator.ingest.apiclient.reposerver;

import com.greenops.workfloworchestrator.datamodel.requests.GetFileRequest;

public interface RepoManagerApi {
    public boolean getFileFromRepo(GetFileRequest getFileRequest, String orgName, String teamName);
}
