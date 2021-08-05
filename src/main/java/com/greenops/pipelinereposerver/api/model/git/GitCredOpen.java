package com.greenops.pipelinereposerver.api.model.git;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;

@JsonDeserialize(as = GitCredOpen.class)
public class GitCredOpen implements GitCred {

    @Override
    public String convertGitCredToString(String gitRepoLink) {
        return gitRepoLink;
    }
}
