package com.greenops.workfloworchestrator.datamodel.mixin.pipelinedata;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.Map;

public abstract class TestMixin {

    @JsonProperty(value = "path")
    String path;

    @JsonProperty(value = "in_application_pod")
    boolean executeInApplicationPod;

    @JsonProperty(value = "before")
    boolean executeBeforeDeployment;

    @JsonProperty(value = "variables")
    Map<String, String> variables;

    @JsonCreator
    public TestMixin(@JsonProperty(value = "path") String path,
              @JsonProperty(value = "in_application_pod") boolean executeInApplicationPod,
              @JsonProperty(value = "before") boolean executeBeforeDeployment,
              @JsonProperty(value = "variables") Map<String, String> variables) {
    }

    @JsonIgnore
    abstract String getPath();

    @JsonIgnore
    abstract boolean shouldExecuteInPod();

    @JsonIgnore
    abstract boolean shouldExecuteBefore();

    @JsonIgnore
    abstract Map<String, String> getVariables();

}
