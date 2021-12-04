package com.greenops.util.datamodel.mixin.clientmessages;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class ClientDeleteByGvkRequestMixin {

    @JsonProperty("orgName")
    String orgName;
    @JsonProperty("teamName")
    String teamName;
    @JsonProperty("pipelineName")
    String pipelineName;
    @JsonProperty("pipelineUvn")
    String uvn;
    @JsonProperty("stepName")
    String stepName;
    @JsonProperty("type")
    String type;
    @JsonProperty("resourceName")
    String resourceName;
    @JsonProperty("resourceNamespace")
    String resourceNamespace;
    @JsonProperty("group")
    String group;
    @JsonProperty("version")
    String version;
    @JsonProperty("kind")
    String kind;
    @JsonProperty("finalTry")
    boolean finalTry;

    @JsonCreator
    public ClientDeleteByGvkRequestMixin(@JsonProperty("orgName") String orgName,
                                         @JsonProperty("teamName") String teamName,
                                         @JsonProperty("pipelineName") String pipelineName,
                                         @JsonProperty("pipelineUvn") String uvn,
                                         @JsonProperty("stepName") String stepName,
                                         @JsonProperty("type") String type,
                                         @JsonProperty("resourceName") String resourceName,
                                         @JsonProperty("resourceNamespace") String resourceNamespace,
                                         @JsonProperty("group") String group,
                                         @JsonProperty("version") String version,
                                         @JsonProperty("kind") String kind) {
    }
}
