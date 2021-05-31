package com.greenops.pipelinereposerver.repomanager;

import com.greenops.pipelinereposerver.api.model.git.GitRepoSchema;

public interface RepoManager {

    boolean clone(GitRepoSchema gitRepoSchema);
    boolean delete(GitRepoSchema gitRepoSchema);
    //TODO: soon update needs to be added in
}
