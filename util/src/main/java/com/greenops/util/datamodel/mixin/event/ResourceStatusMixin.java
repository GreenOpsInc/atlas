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
    @JsonProperty(value = "Group")
    String group;
    @JsonProperty(value = "Version")
    String version;
    @JsonProperty(value = "Kind")
    String kind;

    @JsonCreator
    public ResourceStatusMixin(@JsonProperty(value = "resourceName") String resourceName,
                               @JsonProperty(value = "resourceNamespace") String resourceNamespace,
                               @JsonProperty(value = "healthStatus") String healthStatus,
                               @JsonProperty(value = "syncStatus") String syncStatus,
                               @JsonProperty(value = "Group") String group,
                               @JsonProperty(value = "Version") String version,
                               @JsonProperty(value = "Kind") String kind) {
    }
}
