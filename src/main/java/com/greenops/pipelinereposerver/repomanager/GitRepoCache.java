package com.greenops.pipelinereposerver.repomanager;

import com.greenops.pipelinereposerver.api.model.git.GitRepoSchema;

import java.util.ArrayList;
import java.util.List;

public class GitRepoCache {

    private String rootCommitHash;
    private List<String> commitHashHistory;
    private GitRepoSchema gitRepoSchema;

    GitRepoCache(String rootCommitHash, String commitHash, GitRepoSchema gitRepoSchema) {
        this.rootCommitHash = rootCommitHash;
        this.commitHashHistory = new ArrayList<>();
        this.commitHashHistory.add(commitHash);
        this.gitRepoSchema = gitRepoSchema;
    }

    String getRootCommitHash() {
        return rootCommitHash;
    }

    void setRootCommitHash(String rootCommitHash) {
        this.rootCommitHash = rootCommitHash;
    }

    List<String> getCommitHashHistory() {
        return commitHashHistory;
    }

    GitRepoSchema getGitRepoSchema() {
        return gitRepoSchema;
    }

    void addCommitHashToHistory(String commitHash) {
        if (commitHashHistory.size() == 0) {
            commitHashHistory.add(commitHash);
        } else {
            commitHashHistory.add(0, commitHash);
        }
    }
}
