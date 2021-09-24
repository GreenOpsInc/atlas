package com.greenops.util.datamodel.mixin.event;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class ApplicationInfraTriggerEventMixin {

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

    @JsonCreator
    public ApplicationInfraTriggerEventMixin(@JsonProperty(value = "orgName") String orgName,
                                             @JsonProperty(value = "teamName") String teamName,
                                             @JsonProperty(value = "pipelineName") String pipelineName,
                                             @JsonProperty(value = "pipelineUvn") String uvn,
                                             @JsonProperty(value = "stepName") String stepName) {
    }

    @JsonIgnore
    abstract String getPipelineUvn();
}
