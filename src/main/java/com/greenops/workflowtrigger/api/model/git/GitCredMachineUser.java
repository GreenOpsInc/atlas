package com.greenops.workflowtrigger.api.model.git;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;

@JsonDeserialize(as = GitCredMachineUser.class)
public class GitCredMachineUser implements GitCred {

    private String username;
    private String password;

    public GitCredMachineUser(String username, String password) {
        this.username = username;
        this.password = password;
    }
}
