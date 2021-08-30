package com.greenops.util.datamodel.mixin.clientmessages;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class ResourceGvkMixin {

    @JsonProperty(value = "resourceName")
    String resourceName;
    @JsonProperty(value = "resourceNamespace")
    String resourceNamespace;
    @JsonProperty(value = "group")
    String group;
    @JsonProperty(value = "version")
    String version;
    @JsonProperty(value = "kind")
    String kind;

    @JsonCreator
    public ResourceGvkMixin(@JsonProperty(value = "resourceName") String resourceName,
                            @JsonProperty(value = "resourceNamespace") String resourceNamespace,
                            @JsonProperty(value = "group") String group,
                            @JsonProperty(value = "version") String version,
                            @JsonProperty(value = "kind") String kind) {
    }
}
