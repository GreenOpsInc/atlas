package com.greenops.pipelinereposerver.api.model.mixin.request;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class GetFileRequestMixin {

    @JsonProperty("gitRepoUrl")
    String gitRepoUrl;
    @JsonProperty("filename")
    String filename;

    @JsonCreator
    public GetFileRequestMixin(@JsonProperty("gitRepoUrl") String gitRepoUrl, @JsonProperty("filename") String filename) {
    }

}
