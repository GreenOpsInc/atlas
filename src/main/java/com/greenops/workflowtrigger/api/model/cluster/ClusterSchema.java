package com.greenops.workflowtrigger.api.model.cluster;


public class ClusterSchema {
    private String clusterIP;
    private int exposedPort;
    private String clusterName;

    public ClusterSchema(String clusterIP, int exposedPort, String clusterName) {
        this.clusterIP = clusterIP;
        this.exposedPort = exposedPort;
        this.clusterName = clusterName;
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
}

