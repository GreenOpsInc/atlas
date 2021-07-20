package com.greenops.workflowtrigger.api.model.git;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;

@JsonDeserialize(as = GitCredToken.class)
public class GitCredToken implements GitCred {

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
    public boolean equals(Object o){
        if(o instanceof GitCredToken){
            return token.equals(((GitCredToken) o).token);
        }
        return false;
    }
}
