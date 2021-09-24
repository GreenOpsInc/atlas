package com.greenops.util.datamodel.mixin.clientmessages;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class ClientDeployRequestMixin {

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
    @JsonProperty("responseEventType")
    String responseEventType;
    @JsonProperty("deployType")
    String deployType;
    @JsonProperty("revisionHash")
    String revisionHash;
    @JsonProperty("payload")
    String payload;

    @JsonCreator
    public ClientDeployRequestMixin(@JsonProperty("orgName") String orgName,
                                    @JsonProperty("teamName") String teamName,
                                    @JsonProperty("pipelineName") String pipelineName,
                                    @JsonProperty("pipelineUvn") String uvn,
                                    @JsonProperty("stepName") String stepName,
                                    @JsonProperty("responseEventType") String responseEventType,
                                    @JsonProperty("deployType") String deployType,
                                    @JsonProperty("revisionHash") String revisionHash,
                                    @JsonProperty("payload") String payload) {
    }
}
