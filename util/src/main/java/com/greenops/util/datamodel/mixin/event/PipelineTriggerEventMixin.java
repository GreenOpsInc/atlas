package com.greenops.util.datamodel.mixin.event;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class PipelineTriggerEventMixin extends EventMixin {

    @JsonProperty(value = "revisionHash")
    String revisionHash;

    @JsonCreator
    public PipelineTriggerEventMixin(@JsonProperty(value = "orgName") String orgName,
                                     @JsonProperty(value = "teamName") String teamName,
                                     @JsonProperty(value = "pipelineName") String pipelineName,
                                     @JsonProperty(value = "pipelineUvn") String pipelineUvn,
                                     @JsonProperty(value = "stepName") String stepName,
                                     @JsonProperty(value = "revisionHash") String revisionHash) {
    }
}
