package com.greenops.util.datamodel.mixin.event;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class ApplicationInfraTriggerEventMixin extends EventMixin {

    @JsonCreator
    public ApplicationInfraTriggerEventMixin(@JsonProperty(value = "orgName") String orgName,
                                             @JsonProperty(value = "teamName") String teamName,
                                             @JsonProperty(value = "pipelineName") String pipelineName,
                                             @JsonProperty(value = "pipelineUvn") String pipelineUvn,
                                             @JsonProperty(value = "stepName") String stepName) {
    }
}
