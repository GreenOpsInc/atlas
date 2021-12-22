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
    @JsonProperty(value = "stepName")
    String stepName;
    @JsonProperty(value = "revisionHash")
    String revisionHash;
    @JsonProperty(value = "pathToRoot")
    String pathToRoot;

    @JsonCreator
    public PipelineTriggerEventMixin(@JsonProperty(value = "orgName") String orgName,
                                     @JsonProperty(value = "teamName") String teamName,
                                     @JsonProperty(value = "pipelineName") String pipelineName,
                                     @JsonProperty(value = "pipelineUvn") String uvn,
                                     @JsonProperty(value = "stepName") String stepName,
                                     @JsonProperty(value = "revisionHash") String revisionHash,
                                     @JsonProperty(value = "pathToRoot") String pathToRoot) {
    }

    @JsonIgnore
    abstract String getPipelineUvn();
}
