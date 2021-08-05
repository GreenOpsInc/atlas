package com.greenops.util.datamodel.mixin.git;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;

public abstract class GitCredOpenMixin {

    @JsonCreator
    public GitCredOpenMixin() {
    }

    @JsonIgnore
    abstract String convertGitCredToString(String gitRepoLink);

    @JsonIgnore
    abstract void hide();
}
