package com.greenops.util.datamodel.mixin.pipelinedata;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.Map;

public abstract class ArgoWorkflowTaskMixin {

    @JsonProperty(value = "path")
    String path;

    @JsonProperty(value = "before")
    boolean executeBeforeDeployment;

    @JsonProperty(value = "variables")
    Map<String, String> variables;

    @JsonCreator
    public ArgoWorkflowTaskMixin(@JsonProperty(value = "path") String path,
                                 @JsonProperty(value = "before") boolean executeBeforeDeployment,
                                 @JsonProperty(value = "variables") Map<String, String> variables) {
    }

    @JsonIgnore
    abstract String getPath();

    @JsonIgnore
    abstract boolean shouldExecuteBefore();

    @JsonIgnore
    abstract Map<String, String> getVariables();
}
