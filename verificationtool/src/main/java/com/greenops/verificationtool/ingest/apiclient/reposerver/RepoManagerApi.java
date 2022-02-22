package com.greenops.verificationtool.ingest.apiclient.reposerver;

import com.greenops.util.datamodel.request.GetFileRequest;

public interface RepoManagerApi {

    public String getFileFromRepo(GetFileRequest getFileRequest, String orgName, String teamName);
}
