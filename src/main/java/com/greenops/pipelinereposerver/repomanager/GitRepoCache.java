package com.greenops.pipelinereposerver.repomanager;

import com.greenops.pipelinereposerver.api.model.git.GitRepoSchema;

import java.util.ArrayList;
import java.util.List;

public class GitRepoCache {

    private List<String> commitHashHistory;
    private GitRepoSchema gitRepoSchema;

    GitRepoCache(String commitHash, GitRepoSchema gitRepoSchema) {
        this.commitHashHistory = new ArrayList<>();
        this.commitHashHistory.add(commitHash);
        this.gitRepoSchema = gitRepoSchema;
    }

    List<String> getCommitHashHistory() {
        return commitHashHistory;
    }

    GitRepoSchema getGitRepoSchema() {
        return gitRepoSchema;
    }

    void addCommitHashToHistory(String commitHash) {
        commitHashHistory.add(commitHash);
    }
}
