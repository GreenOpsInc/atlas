package com.greenops.workfloworchestrator.datamodel.mixin.requests;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.List;
import java.util.Map;

public abstract class KubernetesCreationRequestMixin {

    @JsonProperty(value = "kind")
    String kind;

    @JsonProperty(value = "objectName")
    String objectName;

    @JsonProperty(value = "namespace")
    String namespace;

    @JsonProperty(value = "imageName")
    String imageName;

    @JsonProperty(value = "command")
    List<String> command;

    @JsonProperty(value = "args")
    List<String> args;

    @JsonProperty(value = "configPayload")
    String configPayload;

    @JsonProperty(value = "volumeFilename")
    String volumeFilename;

    @JsonProperty(value = "volumePayload")
    String volumePayload;

    @JsonProperty(value = "variables")
    Map<String, String> variables;

    @JsonCreator
    public KubernetesCreationRequestMixin(@JsonProperty(value = "kind") String kind,
                                          @JsonProperty(value = "objectName") String objectName,
                                          @JsonProperty(value = "namespace") String namespace,
                                          @JsonProperty(value = "imageName") String imageName,
                                          @JsonProperty(value = "command") List<String> command,
                                          @JsonProperty(value = "args") List<String> args,
                                          @JsonProperty(value = "configPayload") String configPayload,
                                          @JsonProperty(value = "volumeFilename") String volumeFilename,
                                          @JsonProperty(value = "volumePayload") String volumePayload,
                                          @JsonProperty(value = "variables") Map<String, String> variables) {
    }
}
