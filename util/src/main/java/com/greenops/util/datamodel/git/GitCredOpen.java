package com.greenops.util.datamodel.git;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;

@JsonDeserialize(as = GitCredOpen.class)
public class GitCredOpen implements GitCred, GitCredAccessible {

    @Override
    public String convertGitCredToString(String gitRepoLink) {
        return gitRepoLink;
    }

    @Override
    public void hide() {
    }

    @Override
    public boolean equals(Object o) {
        return o instanceof GitCredOpen;
    }
}
