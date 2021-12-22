package com.greenops.util.datamodel.mixin.clientmessages;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class AggregateRequestMixin {

    @JsonProperty("orgName")
    private String orgName;
    @JsonProperty("teamName")
    private String teamName;
    @JsonProperty("pipelineName")
    private String pipelineName;
    @JsonProperty("pipelineUvn")
    private String uvn;
    @JsonProperty("stepName")
    private String stepName;
    @JsonProperty(value = "clusterName")
    private String clusterName;
    @JsonProperty(value = "namespace")
    private String namespace;
    @JsonProperty(value = "requestId")
    private String requestId;
    @JsonProperty("finalTry")
    private boolean finalTry;


    @JsonCreator
    public AggregateRequestMixin(@JsonProperty("orgName") String orgName,
                                @JsonProperty("teamName") String teamName,
                                @JsonProperty("pipelineName") String pipelineName,
                                @JsonProperty("pipelineUvn") String uvn,
                                @JsonProperty("stepName") String stepName,
                                @JsonProperty(value = "clusterName") String clusterName,
                                @JsonProperty(value = "namespace") String namespace,
                                @JsonProperty(value = "requestId") String requestId) {
    }

}
