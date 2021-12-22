package com.greenops.util.datamodel.cluster;

public class NoDeployInfo {

    private String name;
    private String namespace;
    private String reason;

    public NoDeployInfo(String name, String namespace, String reason) {
        this.name = name;
        this.namespace = namespace;
        this.reason = reason;
    }

    public String getName() {
        return name;
    }

    public String getNamespace() {
        return namespace;
    }

    public String getReason() {
        return reason;
    }
}
