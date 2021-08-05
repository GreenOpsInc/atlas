package com.greenops.pipelinereposerver.repomanager;

import com.greenops.util.datamodel.git.GitRepoSchema;

public class GitRepoCache {

    private String rootCommitHash;
    private String currentCommitHash;
    private GitRepoSchema gitRepoSchema;

    GitRepoCache(String rootCommitHash, String commitHash, GitRepoSchema gitRepoSchema) {
        this.rootCommitHash = rootCommitHash;
        this.currentCommitHash = commitHash;
        this.gitRepoSchema = gitRepoSchema;
    }

    String getRootCommitHash() {
        return rootCommitHash;
    }

    void setRootCommitHash(String rootCommitHash) {
        this.rootCommitHash = rootCommitHash;
    }

    String getCurrentCommitHash() {
        return currentCommitHash;
    }

    GitRepoSchema getGitRepoSchema() {
        return gitRepoSchema;
    }

    void setCurrentCommitHash(String commitHash) {
        currentCommitHash = commitHash;
    }
}
