package com.greenops.workfloworchestrator.datamodel.git;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;

@JsonDeserialize(as = GitCredMachineUser.class)
public class GitCredMachineUser implements GitCred {

    private String username;
    private String password;

    /**
     * This is a private constructor meant solely to let Jackson
     * deserialize objects with a singular instance variable, and
     * to keep the code consistent across all of different types of these
     * objects. A Dev should not be calling this at all.
     */
    private GitCredMachineUser() {
    }

    public GitCredMachineUser(String username, String password) {
        this.username = username;
        this.password = password;
    }
}
