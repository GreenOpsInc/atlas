package com.greenops.workfloworchestrator.datamodel.mixin.event;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class ClientCompletionEventMixin {

    @JsonProperty(value = "healthStatus")
    String healthStatus;
    @JsonProperty(value = "orgName")
    String orgName;
    @JsonProperty(value = "teamName")
    String teamName;
    @JsonProperty(value = "pipelineName")
    String pipelineName;
    @JsonProperty(value = "stepName")
    String stepName;
    @JsonProperty(value = "argoName")
    String argoName;
    @JsonProperty(value = "operation")
    String operation;
    @JsonProperty(value = "project")
    String project;
    @JsonProperty(value = "repo")
    String repo;

    @JsonCreator
    public ClientCompletionEventMixin(@JsonProperty(value = "healthStatus") String healthStatus,
                                      @JsonProperty(value = "orgName") String orgName,
                                      @JsonProperty(value = "teamName") String teamName,
                                      @JsonProperty(value = "pipelineName") String pipelineName,
                                      @JsonProperty(value = "stepName") String stepName,
                                      @JsonProperty(value = "argoName") String argoName,
                                      @JsonProperty(value = "operation") String operation,
                                      @JsonProperty(value = "project") String project,
                                      @JsonProperty(value = "repo") String repo) {
    }
}