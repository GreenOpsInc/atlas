package com.greenops.verificationtool.datamodel.mixin.pipelinedata;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.Map;

public abstract class CustomJobTestMixin {

    @JsonProperty(value = "path")
    String path;

    @JsonProperty(value = "before")
    boolean executeBeforeDeployment;

    @JsonProperty(value = "variables")
    Map<String, String> variables;

    @JsonCreator
    public CustomJobTestMixin(@JsonProperty(value = "path") String path,
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
