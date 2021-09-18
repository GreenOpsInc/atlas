package com.greenops.util.datamodel.mixin.pipelinestatus;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.greenops.util.datamodel.auditlog.DeploymentLog;
import com.greenops.util.datamodel.pipelinestatus.FailedStep;

import java.util.List;

public abstract class PipelineStatusMixin {

    @JsonProperty("progressingSteps")
    List<String> progressingSteps;

    @JsonProperty("stable")
    boolean stable;

    @JsonProperty("complete")
    boolean complete;

    @JsonProperty("cancelled")
    boolean cancelled;

    @JsonProperty("failedSteps")
    List<FailedStep> failedSteps;

    @JsonCreator
    public PipelineStatusMixin(@JsonProperty("progressingSteps") List<String> progressingSteps,
                               @JsonProperty("stable") boolean stable,
                               @JsonProperty("complete") boolean complete,
                               @JsonProperty("cancelled") boolean cancelled,
                               @JsonProperty("failedSteps") List<FailedStep> failedSteps) {
    }

    @JsonIgnore
    abstract void markCancelled();

    @JsonIgnore
    abstract void markIncomplete();

    @JsonIgnore
    abstract void addLatestLog(DeploymentLog log);

    @JsonIgnore
    abstract void addLatestDeploymentLog(DeploymentLog log, String step);

    @JsonIgnore
    abstract void addProgressingStep(String stepName);

    @JsonIgnore
    abstract void addFailedDeploymentLog(DeploymentLog log, String step);
}
