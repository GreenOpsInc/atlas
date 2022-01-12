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

    @JsonCreator
    public PipelineCompletionEventMixin(@JsonProperty(value = "orgName") String orgName,
                                        @JsonProperty(value = "teamName") String teamName,
                                        @JsonProperty(value = "pipelineName") String pipelineName,
                                        @JsonProperty(value = "pipelineUvn") String uvn) {
    }

    @JsonIgnore
    abstract String getPipelineUvn();
}
