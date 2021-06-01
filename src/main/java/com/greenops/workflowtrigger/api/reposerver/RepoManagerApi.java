package com.greenops.workflowtrigger.api.reposerver;

import com.greenops.workflowtrigger.api.model.git.GitRepoSchema;

public interface RepoManagerApi {

    boolean cloneRepo(GitRepoSchema gitRepoSchema);
    boolean deleteRepo(GitRepoSchema gitRepoSchema);
}
