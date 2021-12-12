package com.greenops.util.datamodel.cluster;

public class NoDeployInfo {

    private String name;
    private String reason;

    public NoDeployInfo(String name, String reason) {
        this.name = name;
        this.reason = reason;
    }

    public String getName() {
        return name;
    }

    public String getReason() {
        return reason;
    }
}
