package com.greenops.workfloworchestrator.datamodel.mixin.pipelinedata;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.List;

public abstract class ArgoWorkflowTaskMixin {

    @JsonProperty(value = "path")
    String path;

    @JsonProperty(value = "before")
    boolean executeBeforeDeployment;

    @JsonProperty(value = "variables")
    List<Object> variables;

    @JsonCreator
    public ArgoWorkflowTaskMixin(@JsonProperty(value = "path") String path,
                                 @JsonProperty(value = "before") boolean executeBeforeDeployment,
                                 @JsonProperty(value = "variables") List<Object> variables) {
    }

    @JsonIgnore
    abstract String getPath();

    @JsonIgnore
    abstract boolean shouldExecuteBefore();

    @JsonIgnore
    abstract List<Object> getVariables();
}
