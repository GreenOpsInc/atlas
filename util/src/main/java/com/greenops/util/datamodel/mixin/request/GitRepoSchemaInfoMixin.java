package com.greenops.util.datamodel.mixin.request;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class GitRepoSchemaInfoMixin {

    @JsonProperty("gitRepoUrl")
    String gitRepo;
    @JsonProperty("pathToRoot")
    String pathToRoot;

    @JsonCreator
    public GitRepoSchemaInfoMixin(@JsonProperty("gitRepoUrl") String gitRepo, @JsonProperty("pathToRoot") String pathToRoot) {
    }
}
