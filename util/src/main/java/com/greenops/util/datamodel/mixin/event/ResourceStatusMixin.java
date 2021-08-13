package com.greenops.util.datamodel.mixin.event;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class ResourceStatusMixin {

    @JsonProperty(value = "resourceName")
    String resourceName;
    @JsonProperty(value = "resourceNamespace")
    String resourceNamespace;
    @JsonProperty(value = "healthStatus")
    String healthStatus;
    @JsonProperty(value = "syncStatus")
    String syncStatus;
    @JsonProperty(value = "group")
    String group;
    @JsonProperty(value = "version")
    String version;
    @JsonProperty(value = "kind")
    String kind;

    @JsonCreator
    public ResourceStatusMixin(@JsonProperty(value = "resourceName") String resourceName,
                               @JsonProperty(value = "resourceNamespace") String resourceNamespace,
                               @JsonProperty(value = "healthStatus") String healthStatus,
                               @JsonProperty(value = "syncStatus") String syncStatus,
                               @JsonProperty(value = "group") String group,
                               @JsonProperty(value = "version") String version,
                               @JsonProperty(value = "kind") String kind) {
    }
}
