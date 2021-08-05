package com.greenops.util.datamodel.mixin.request;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class GetFileRequestMixin {

    @JsonProperty("gitRepoUrl")
    String gitRepoUrl;
    @JsonProperty("filename")
    String filename;
    @JsonProperty("gitCommitHash")
    String gitCommitHash;

    @JsonCreator
    public GetFileRequestMixin(@JsonProperty("gitRepoUrl") String gitRepoUrl, @JsonProperty("filename") String filename, @JsonProperty("gitCommitHash") String gitCommitHash) {
    }

}
