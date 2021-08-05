package com.greenops.util.datamodel.git;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;

import static com.greenops.util.datamodel.git.GitCredAccessible.SECURE_GIT_URL_PREFIX;

@JsonDeserialize(as = GitCredMachineUser.class)
public class GitCredMachineUser implements GitCred, GitCredAccessible {

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

    @Override
    public String convertGitCredToString(String gitRepoLink) {
        var splitLink = gitRepoLink.split(SECURE_GIT_URL_PREFIX);
        splitLink[1] = username + ':' + password + '@' + splitLink[1];
        return SECURE_GIT_URL_PREFIX + splitLink[1];
    }

    @Override
    public void hide() {
        username = HIDDEN;
        password = HIDDEN;
    }

    @Override
    public boolean equals(Object o) {
        if (o instanceof GitCredMachineUser) {
            return ((GitCredMachineUser) o).password.equals(password) && ((GitCredMachineUser) o).username.equals(username);
        }
        return false;
    }
}
