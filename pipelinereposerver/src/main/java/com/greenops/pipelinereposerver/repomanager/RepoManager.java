package com.greenops.pipelinereposerver.repomanager;

import com.greenops.util.datamodel.git.GitRepoSchema;

import java.util.Set;

public interface RepoManager {

    String getOrgName();

    Set<GitRepoCache> getGitRepos();

    boolean clone(GitRepoSchema gitRepoSchema);

    boolean update(GitRepoSchema gitRepoSchema);

    boolean delete(GitRepoSchema gitRepoSchema);

    String getYamlFileContents(String gitRepoUrl, String filename);

    boolean sync(GitRepoSchema gitRepoSchema);

    boolean resetToVersion(String gitCommit, String gitRepoUrl);

    String getLatestCommitFromCache(String gitRepoUrl);

    boolean containsGitRepoSchema(GitRepoSchema gitRepoSchema);

    public String getRootCommit(String gitRepoUrl);

    String getCurrentCommit(String gitRepoURL);
}
