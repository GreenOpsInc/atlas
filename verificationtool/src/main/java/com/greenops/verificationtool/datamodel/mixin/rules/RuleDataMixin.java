package com.greenops.verificationtool.datamodel.mixin.rules;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.greenops.util.datamodel.auditlog.Log;
import com.greenops.util.datamodel.pipelinestatus.PipelineStatus;

public abstract class RuleDataMixin {
    @JsonProperty(value = "stepName")
    String stepName;
    @JsonProperty(value = "eventType")
    String eventType;
    @JsonProperty(value = "pipelineStatus")
    PipelineStatus pipelineStatus;
    @JsonProperty(value = "stepStatus")
    Log log;

    @JsonCreator
    public RuleDataMixin(@JsonProperty(value = "stepName") String stepName,
                         @JsonProperty(value = "eventType") String eventType,
                         @JsonProperty(value = "pipelineStatus") PipelineStatus pipelineStatus,
                         @JsonProperty(value = "stepStatus") Log log) {

    }
}
