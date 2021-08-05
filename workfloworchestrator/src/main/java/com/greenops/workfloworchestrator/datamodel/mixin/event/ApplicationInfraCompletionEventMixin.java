package com.greenops.workfloworchestrator.datamodel.mixin.event;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class ApplicationInfraCompletionEventMixin {

    @JsonProperty(value = "orgName")
    String orgName;
    @JsonProperty(value = "teamName")
    String teamName;
    @JsonProperty(value = "pipelineName")
    String pipelineName;
    @JsonProperty(value = "stepName")
    String stepName;

    @JsonCreator
    public ApplicationInfraCompletionEventMixin(@JsonProperty(value = "orgName") String orgName,
                                                @JsonProperty(value = "teamName") String teamName,
                                                @JsonProperty(value = "pipelineName") String pipelineName,
                                                @JsonProperty(value = "stepName") String stepName) {
    }
}
