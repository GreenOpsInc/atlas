package com.greenops.pipelinereposerver.api.model.git;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;

@JsonDeserialize(as = GitCredMachineUser.class)
public class GitCredMachineUser implements GitCred {

    private final String username;
    private final String password;

    public GitCredMachineUser(String username, String password) {
        this.username = username;
        this.password = password;
    }

    @Override
    public String convertGitCredToString(String gitRepoLink) {
        var splitLink = gitRepoLink.split(SECURE_GIT_URL_PREFIX);
        splitLink[1] = username + ':' + password + '@' + splitLink[1];
        return SECURE_GIT_URL_PREFIX + splitLink[1];
    }
}
