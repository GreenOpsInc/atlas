package com.greenops.util.datamodel.mixin.event;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.greenops.util.datamodel.request.DeployResponse;

public abstract class FailureEventMixin {

    @JsonProperty(value = "orgName")
    String orgName;
    @JsonProperty(value = "teamName")
    String teamName;
    @JsonProperty(value = "pipelineName")
    String pipelineName;
    @JsonProperty(value = "pipelineUvn")
    String uvn;
    @JsonProperty(value = "stepName")
    String stepName;
    @JsonProperty(value = "deployResponse")
    DeployResponse deployResponse;
    @JsonProperty(value = "statusCode")
    String statusCode;
    @JsonProperty(value = "error")
    String error;

    @JsonCreator
    public FailureEventMixin(@JsonProperty(value = "orgName") String orgName,
                             @JsonProperty(value = "teamName") String teamName,
                             @JsonProperty(value = "pipelineName") String pipelineName,
                             @JsonProperty(value = "pipelineUvn") String uvn,
                             @JsonProperty(value = "stepName") String stepName,
                             @JsonProperty(value = "deployResponse") DeployResponse deployResponse,
                             @JsonProperty(value = "statusCode") String statusCode,
                             @JsonProperty(value = "error") String error) {
    }

    @JsonIgnore
    abstract String getPipelineUvn();
}
