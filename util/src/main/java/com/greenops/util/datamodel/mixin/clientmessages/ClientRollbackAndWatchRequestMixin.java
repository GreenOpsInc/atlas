package com.greenops.util.datamodel.mixin.clientmessages;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class ClientRollbackAndWatchRequestMixin {

    //For rolling back
    @JsonProperty("orgName")
    String orgName;
    @JsonProperty("appName")
    String appName;
    @JsonProperty("revisionHash")
    String revisionHash;
    //For watching
    @JsonProperty("watchType")
    String watchType;
    @JsonProperty("pipelineUvn")
    String uvn;
    @JsonProperty("teamName")
    String teamName;
    @JsonProperty("pipelineName")
    String pipelineName;
    @JsonProperty("stepName")
    String stepName;
    @JsonProperty("finalTry")
    boolean finalTry;

    @JsonCreator
    public ClientRollbackAndWatchRequestMixin(@JsonProperty("orgName") String orgName,
                                              @JsonProperty("pipelineUvn") String uvn,
                                              @JsonProperty("appName") String appName,
                                              @JsonProperty("revisionHash") String revisionHash,
                                              @JsonProperty("watchType") String watchType,
                                              @JsonProperty("teamName") String teamName,
                                              @JsonProperty("pipelineName") String pipelineName,
                                              @JsonProperty("stepName") String stepName) {
    }
}
