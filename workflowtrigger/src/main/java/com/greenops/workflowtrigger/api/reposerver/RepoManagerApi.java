package com.greenops.workflowtrigger.api.reposerver;

import com.greenops.util.datamodel.git.GitRepoSchema;
import com.greenops.util.datamodel.request.GetFileRequest;

public interface RepoManagerApi {

    boolean cloneRepo(String orgName, GitRepoSchema gitRepoSchema);
    boolean deleteRepo(GitRepoSchema gitRepoSchema);
    boolean sync(GitRepoSchema gitRepoSchema);
    String getFileFromRepo(GetFileRequest getFileRequest, String orgName, String teamName);
}
