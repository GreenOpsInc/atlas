package com.greenops.util.datamodel.mixin.event;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;

// This Event Mixin is only used by verificationtool
public abstract class PipelineCompletionEventMixin {

    @JsonProperty(value = "orgName")
    String orgName;
    @JsonProperty(value = "teamName")
    String teamName;
    @JsonProperty(value = "pipelineName")
    String pipelineName;
    @JsonProperty(value = "pipelineUvn")
    String uvn;
    @JsonProperty(value = "status")
    String status;
    @JsonProperty(value = "stepName")
    String stepName;

    @JsonCreator
    public PipelineCompletionEventMixin(@JsonProperty(value = "orgName") String orgName,
                                        @JsonProperty(value = "teamName") String teamName,
                                        @JsonProperty(value = "pipelineName") String pipelineName,
                                        @JsonProperty(value = "stepName") String stepName,
                                        @JsonProperty(value = "pipelineUvn") String uvn,
                                        @JsonProperty(value = "status") String status) {
    }

    @JsonIgnore
    abstract String getPipelineUvn();
}
