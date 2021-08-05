package com.greenops.util.datamodel.git;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;

import static com.greenops.util.datamodel.git.GitCredAccessible.SECURE_GIT_URL_PREFIX;

@JsonDeserialize(as = GitCredToken.class)
public class GitCredToken implements GitCred, GitCredAccessible {

    private String token;

    /**
     * This is a private constructor meant solely to let Jackson
     * deserialize objects with a singular instance variable, and
     * to keep the code consistent across all of different types of these
     * objects. A Dev should not be calling this at all.
     */
    private GitCredToken() {
    }

    public GitCredToken(String token) {
        this.token = token;
    }

    @Override
    public String convertGitCredToString(String gitRepoLink) {
        var splitLink = gitRepoLink.split(SECURE_GIT_URL_PREFIX);
        splitLink[1] = token + '@' + splitLink[1];
        return SECURE_GIT_URL_PREFIX + splitLink[1];
    }

    @Override
    public void hide() {
        token = HIDDEN;
    }

    @Override
    public boolean equals(Object o) {
        if (o instanceof GitCredToken) {
            return token.equals(((GitCredToken) o).token);
        }
        return false;
    }
}
