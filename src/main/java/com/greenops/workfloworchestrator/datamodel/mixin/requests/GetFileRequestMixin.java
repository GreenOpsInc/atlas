package com.greenops.workfloworchestrator.datamodel.mixin.requests;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class GetFileRequestMixin {

    @JsonProperty(value = "gitRepoUrl")
    String gitRepoUrl;
    @JsonProperty(value = "filename")
    String filename;
    @JsonProperty(value = "gitCommitHash")
    String gitCommitHash;

    @JsonCreator
    public GetFileRequestMixin(@JsonProperty(value = "gitRepoUrl") String gitRepoUrl, @JsonProperty(value = "filename") String filename, @JsonProperty(value = "gitCommitHash") String gitCommitHash) {
    }
}
