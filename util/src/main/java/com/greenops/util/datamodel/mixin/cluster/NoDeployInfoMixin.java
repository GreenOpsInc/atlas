package com.greenops.util.datamodel.mixin.cluster;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class NoDeployInfoMixin {
    @JsonProperty("name")
    String name;

    @JsonProperty("namespace")
    String namespace;

    @JsonProperty("reason")
    String reason;

    @JsonCreator
    public NoDeployInfoMixin(@JsonProperty("name") String name,
                             @JsonProperty("namespace") String namespace,
                             @JsonProperty("reason") String reason) {
    }
}
