package com.greenops.verificationtool.datamodel.mixin.requests;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

public class WatchRequestMixin {

    @JsonProperty(value = "teamName")
    String teamName;

    @JsonProperty(value = "pipelineName")
    String pipelineName;

    @JsonProperty(value = "stepName")
    String stepName;

    @JsonProperty(value = "type")
    String type;

    @JsonProperty(value = "name")
    String name;

    @JsonProperty(value = "namespace")
    String namespace;

    @JsonProperty(value = "testNumber")
    int testNumber;

    @JsonCreator
    public WatchRequestMixin(@JsonProperty(value = "teamName") String teamName,
                             @JsonProperty(value = "pipelineName") String pipelineName,
                             @JsonProperty(value = "stepName") String stepName,
                             @JsonProperty(value = "type") String type,
                             @JsonProperty(value = "name") String name,
                             @JsonProperty(value = "namespace") String namespace,
                             @JsonProperty(value = "testNumber") int testNumber) {
    }
}
