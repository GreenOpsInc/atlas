package com.greenops.util.datamodel.mixin.clientmessages;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class ClientDeployNamedArgoAppAndWatchRequestMixin {

    //For deploying
    @JsonProperty("orgName")
    String orgName;
    @JsonProperty("pipelineUvn")
    String uvn;
    @JsonProperty("type")
    String type;
    @JsonProperty("appName")
    String appName;
    //For watching
    @JsonProperty("watchType")
    String watchType;
    @JsonProperty("teamName")
    String teamName;
    @JsonProperty("pipelineName")
    String pipelineName;
    @JsonProperty("stepName")
    String stepName;
    @JsonProperty("finalTry")
    boolean finalTry;


    @JsonCreator
    public ClientDeployNamedArgoAppAndWatchRequestMixin(@JsonProperty("orgName") String orgName,
                                                        @JsonProperty("pipelineUvn") String uvn,
                                                        @JsonProperty("type") String type,
                                                        @JsonProperty("appName") String appName,
                                                        @JsonProperty("watchType") String watchType,
                                                        @JsonProperty("teamName") String teamName,
                                                        @JsonProperty("pipelineName") String pipelineName,
                                                        @JsonProperty("stepName") String stepName) {
    }
}
