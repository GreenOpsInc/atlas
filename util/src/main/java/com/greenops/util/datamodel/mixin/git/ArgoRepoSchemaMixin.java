package com.greenops.util.datamodel.mixin.git;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class ArgoRepoSchemaMixin {

    @JsonProperty("repoUrl")
    String repoUrl;
    @JsonProperty("targetRevision")
    String targetRevision;
    @JsonProperty("path")
    String path;

    @JsonCreator
    public ArgoRepoSchemaMixin(@JsonProperty("repoUrl") String repoUrl, @JsonProperty("targetRevision") String targetRevision, @JsonProperty("path") String path) {
    }
}
