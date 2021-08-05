package com.greenops.workflowtrigger.api.model.mixin.git;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.greenops.workflowtrigger.api.model.git.GitCred;

public abstract class GitRepoSchemaMixin {

    @JsonProperty("gitRepo")
    String gitRepo;

    @JsonProperty("pathToRoot")
    String pathToRoot;

    @JsonProperty("gitCred")
    GitCred gitCred;

    @JsonCreator
    public GitRepoSchemaMixin(@JsonProperty("gitRepo") String gitRepo, @JsonProperty("pathToRoot") String pathToRoot, @JsonProperty("gitCred") GitCred gitCred) {}
}
