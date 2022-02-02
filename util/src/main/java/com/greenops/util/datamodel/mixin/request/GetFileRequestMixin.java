package com.greenops.util.datamodel.mixin.request;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.greenops.util.datamodel.git.GitRepoSchemaInfo;

public abstract class GetFileRequestMixin {

    @JsonProperty("gitRepoSchemaInfo")
    GitRepoSchemaInfo gitRepoSchemaInfo;
    @JsonProperty("filename")
    String filename;
    @JsonProperty("gitCommitHash")
    String gitCommitHash;

    @JsonCreator
    public GetFileRequestMixin(@JsonProperty("gitRepoSchemaInfo") GitRepoSchemaInfo gitRepoSchemaInfo, @JsonProperty("filename") String filename, @JsonProperty("gitCommitHash") String gitCommitHash) {
    }

}
