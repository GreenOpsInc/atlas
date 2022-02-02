package com.greenops.pipelinereposerver.repomanager;

import com.greenops.util.datamodel.git.GitRepoSchema;

import java.util.Set;

public interface RepoManager {

    String getOrgName();

    Set<GitRepoSchema> getGitRepos();

    boolean clone(GitRepoSchema gitRepoSchema);

    boolean update(GitRepoSchema gitRepoSchema);

    boolean delete(GitRepoSchema gitRepoSchema);

    String getYamlFileContents(String filename, GitRepoSchema gitRepoSchema);

    String sync(GitRepoSchema gitRepoSchema);

    boolean resetToVersion(String gitCommit, GitRepoSchema gitRepoSchema);

    boolean containsGitRepoSchema(GitRepoSchema gitRepoSchema);

    String getCurrentCommit(String gitRepoURL);
}
