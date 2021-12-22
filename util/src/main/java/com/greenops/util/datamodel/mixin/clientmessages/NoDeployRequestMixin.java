package com.greenops.util.datamodel.mixin.clientmessages;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class NoDeployRequestMixin {

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
    @JsonProperty(value = "clusterName")
    String clusterName;
    @JsonProperty(value = "namespace")
    String namespace;
    @JsonProperty(value = "apply")
    boolean apply;
    @JsonProperty(value = "requestId")
    String requestId;
    @JsonProperty("finalTry")
    boolean finalTry;


    @JsonCreator
    public NoDeployRequestMixin(@JsonProperty("orgName") String orgName,
                                @JsonProperty("teamName") String teamName,
                                @JsonProperty("pipelineName") String pipelineName,
                                @JsonProperty("pipelineUvn") String uvn,
                                @JsonProperty("stepName") String stepName,
                                @JsonProperty(value = "clusterName") String clusterName,
                                @JsonProperty(value = "namespace") String namespace,
                                @JsonProperty(value = "apply") boolean apply,
                                @JsonProperty(value = "requestId") String requestId) {
    }

}
