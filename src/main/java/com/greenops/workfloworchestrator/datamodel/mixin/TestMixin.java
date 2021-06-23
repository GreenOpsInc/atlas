package com.greenops.workfloworchestrator.datamodel.mixin;

import com.fasterxml.jackson.annotation.JsonCreator;
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

}
