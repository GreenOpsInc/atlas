package com.greenops.util.datamodel.mixin.cluster;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class ClusterSchemaMixin {
    @JsonProperty("clusterIP")
    String clusterIP;

    @JsonProperty("exposedPort")
    int exposedPort;

    @JsonProperty("clusterName")
    String clusterName;

    @JsonCreator
    public ClusterSchemaMixin(@JsonProperty("clusterIP") String clusterIP, @JsonProperty("exposedPort") int exposedPort, @JsonProperty("clusterName") String clusterName) {
    }

}

