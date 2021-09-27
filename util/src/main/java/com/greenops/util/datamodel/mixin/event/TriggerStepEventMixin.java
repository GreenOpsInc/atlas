package com.greenops.util.datamodel.mixin.event;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class TriggerStepEventMixin {

    @JsonProperty(value = "orgName")
    String orgName;
    @JsonProperty(value = "teamName")
    String teamName;
    @JsonProperty(value = "pipelineName")
    String pipelineName;
    @JsonProperty(value = "stepName")
    String stepName;
    @JsonProperty(value = "pipelineUvn")
    String pipelineUvn;
    @JsonProperty(value = "gitCommitHash")
    String gitCommitHash;
    @JsonProperty(value = "rollback")
    boolean rollback;

    @JsonCreator
    public TriggerStepEventMixin(@JsonProperty(value = "orgName") String orgName,
                                 @JsonProperty(value = "teamName") String teamName,
                                 @JsonProperty(value = "pipelineName") String pipelineName,
                                 @JsonProperty(value = "stepName") String stepName,
                                 @JsonProperty(value = "pipelineUvn") String pipelineUvn,
                                 @JsonProperty(value = "gitCommitHash") String gitCommitHash,
                                 @JsonProperty(value = "rollback") boolean rollback) {
    }

    @JsonIgnore
    abstract boolean isRollback();

    @JsonIgnore
    abstract String getPipelineUvn();
}
