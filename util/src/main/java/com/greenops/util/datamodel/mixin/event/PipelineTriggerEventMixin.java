package com.greenops.util.datamodel.mixin.event;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class PipelineTriggerEventMixin {

    @JsonProperty(value = "orgName")
    String orgName;
    @JsonProperty(value = "teamName")
    String teamName;
    @JsonProperty(value = "pipelineName")
    String pipelineName;
    @JsonProperty(value = "pipelineUvn")
    String uvn;

    @JsonCreator
    public PipelineTriggerEventMixin(@JsonProperty(value = "orgName") String orgName,
                                     @JsonProperty(value = "teamName") String teamName,
                                     @JsonProperty(value = "pipelineName") String pipelineName,
                                     @JsonProperty(value = "pipelineUvn") String uvn) {
    }

    @JsonIgnore
    abstract String getPipelineUvn();
}
