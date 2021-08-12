package com.greenops.util.datamodel.event;

public class ResourceStatus {

    private String resourceName;
    private String resourceNamespace;
    private String healthStatus;
    private String syncStatus;
    private String group;
    private String version;
    private String kind;

    ResourceStatus(String resourceName, String resourceNamespace, String healthStatus, String syncStatus, String group, String version, String kind) {
        this.resourceName = resourceName;
        this.resourceNamespace = resourceNamespace;
        this.healthStatus = healthStatus;
        this.syncStatus = syncStatus;
        this.group = group;
        this.version = version;
        this.kind = kind;
    }

    public String getResourceName() {
        return resourceName;
    }

    public String getResourceNamespace() {
        return resourceNamespace;
    }

    public String getHealthStatus() {
        return healthStatus;
    }

    public String getSyncStatus() {
        return syncStatus;
    }

    public String getGroup() {
        return group;
    }

    public String getVersion() {
        return version;
    }

    public String getKind() {
        return kind;
    }
}
