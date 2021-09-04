package com.greenops.util.datamodel.mixin.clientmessages;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class ClientDeployAndWatchRequestMixin {

    //For deploying
    @JsonProperty("orgName")
    String orgName;
    @JsonProperty("deployType")
    String deployType;
    @JsonProperty("revisionHash")
    String revisionHash;
    @JsonProperty("payload")
    String payload;
    //For watching
    @JsonProperty("watchType")
    String watchType;
    @JsonProperty("teamName")
    String teamName;
    @JsonProperty("pipelineName")
    String pipelineName;
    @JsonProperty("stepName")
    String stepName;
    @JsonProperty("testNumber")
    int testNumber;

    @JsonCreator
    public ClientDeployAndWatchRequestMixin(@JsonProperty("orgName") String orgName,
                                            @JsonProperty("deployType") String deployType,
                                            @JsonProperty("revisionHash") String revisionHash,
                                            @JsonProperty("payload") String payload,
                                            @JsonProperty("watchType") String watchType,
                                            @JsonProperty("teamName") String teamName,
                                            @JsonProperty("pipelineName") String pipelineName,
                                            @JsonProperty("stepName") String stepName,
                                            @JsonProperty("testNumber") int testNumber) {
    }
}
