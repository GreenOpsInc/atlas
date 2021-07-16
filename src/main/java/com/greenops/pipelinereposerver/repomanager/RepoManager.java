package com.greenops.pipelinereposerver.repomanager;

import com.greenops.pipelinereposerver.api.model.git.GitRepoSchema;

public interface RepoManager {

    String getOrgName();
    boolean clone(GitRepoSchema gitRepoSchema);
    boolean update(GitRepoSchema gitRepoSchema);
    boolean delete(GitRepoSchema gitRepoSchema);
    String getYamlFileContents(String gitRepoUrl, String filename);
    boolean sync(GitRepoSchema gitRepoSchema);
    boolean resetToVersion(String gitCommit, String gitRepoUrl);
    String getLatestCommitFromCache(String gitRepoUrl);
    boolean containsGitRepoSchema(GitRepoSchema gitRepoSchema);
}
