package com.greenops.util.datamodel.cluster;


public class ClusterSchema {
    private String clusterIP;
    private int exposedPort;
    private String clusterName;
    private NoDeployInfo noDeployInfo;

    public ClusterSchema(String clusterIP, int exposedPort, String clusterName, NoDeployInfo noDeployInfo) {
        this.clusterIP = clusterIP;
        this.exposedPort = exposedPort;
        this.clusterName = clusterName;
        this.noDeployInfo = noDeployInfo;
    }

    public String getClusterIP() {
        return clusterIP;
    }

    public int getExposedPort() {
        return exposedPort;
    }

    public String getClusterName() {
        return clusterName;
    }

    public void setClusterIP(String clusterIP) {
        this.clusterIP = clusterIP;
    }

    public void setClusterName(String clusterName) {
        this.clusterName = clusterName;
    }

    public NoDeployInfo getNoDeployInfo() {
        return noDeployInfo;
    }
}

