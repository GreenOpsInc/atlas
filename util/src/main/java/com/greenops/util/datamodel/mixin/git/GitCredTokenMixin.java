package com.greenops.util.datamodel.mixin.git;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class GitCredTokenMixin {
    @JsonProperty(value = "token")
    String token;

    @JsonCreator
    public GitCredTokenMixin(@JsonProperty(value = "token") String token) {
    }

    @JsonIgnore
    abstract String convertGitCredToString(String gitRepoLink);

    @JsonIgnore
    abstract void hide();
}
