package com.greenops.pipelinereposerver.repomanager;

import com.greenops.pipelinereposerver.api.model.git.GitRepoSchema;

public interface RepoManager {

    boolean clone(GitRepoSchema gitRepoSchema);
    boolean update(GitRepoSchema gitRepoSchema);
    boolean delete(GitRepoSchema gitRepoSchema);
    boolean sync(GitRepoSchema gitRepoSchema);
    boolean containsGitRepoSchema(GitRepoSchema gitRepoSchema);
}
