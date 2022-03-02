package com.greenops.workfloworchestrator.datamodel.mixin.pipelinedata;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.List;
import java.util.Map;

public abstract class InjectScriptTestMixin {

    @JsonProperty(value = "path")
    String path;

    @JsonProperty(value = "image")
    String image;

    @JsonProperty(value = "namespace")
    String namespace;

    @JsonProperty(value = "commands")
    List<String> commands;

    @JsonProperty(value = "arguments")
    List<String> arguments;

    @JsonProperty(value = "in_application_pod")
    boolean executeInApplicationPod;

    @JsonProperty(value = "before")
    boolean executeBeforeDeployment;

    @JsonProperty(value = "variables")
    List<Object> variables;


    @JsonCreator
    public InjectScriptTestMixin(@JsonProperty(value = "path") String path,
                                 @JsonProperty(value = "image") String image,
                                 @JsonProperty(value = "namespace") String namespace,
                                 @JsonProperty(value = "commands") List<String> commands,
                                 @JsonProperty(value = "arguments") List<String> arguments,
                                 @JsonProperty(value = "in_application_pod") boolean executeInApplicationPod,
                                 @JsonProperty(value = "before") boolean executeBeforeDeployment,
                                 @JsonProperty(value = "variables") List<Object> variables) {
    }

    @JsonIgnore
    abstract String getPath();

    @JsonIgnore
    abstract boolean shouldExecuteInPod();

    @JsonIgnore
    abstract boolean shouldExecuteBefore();

    @JsonIgnore
    abstract List<Object> getVariables();

    @JsonIgnore
    abstract String getImage();

    @JsonIgnore
    abstract String getNamespace();

}
