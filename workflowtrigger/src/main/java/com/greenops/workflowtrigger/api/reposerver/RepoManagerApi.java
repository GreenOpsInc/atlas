package com.greenops.workflowtrigger.api.reposerver;

import com.greenops.util.datamodel.git.GitRepoSchema;

public interface RepoManagerApi {

    boolean cloneRepo(GitRepoSchema gitRepoSchema);
    boolean deleteRepo(GitRepoSchema gitRepoSchema);
    boolean sync(GitRepoSchema gitRepoSchema);
}
