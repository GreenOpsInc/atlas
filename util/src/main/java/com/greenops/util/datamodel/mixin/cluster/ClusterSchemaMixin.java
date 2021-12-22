package com.greenops.util.datamodel.mixin.cluster;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.greenops.util.datamodel.cluster.NoDeployInfo;

public abstract class ClusterSchemaMixin {
    @JsonProperty("clusterIP")
    String clusterIP;

    @JsonProperty("exposedPort")
    int exposedPort;

    @JsonProperty("clusterName")
    String clusterName;

    @JsonProperty("noDeploy")
    NoDeployInfo noDeployInfo;

    @JsonCreator
    public ClusterSchemaMixin(@JsonProperty("clusterIP") String clusterIP, @JsonProperty("exposedPort") int exposedPort, @JsonProperty("clusterName") String clusterName, @JsonProperty("noDeploy") NoDeployInfo noDeployInfo) {
    }

}

