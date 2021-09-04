package com.greenops.util.datamodel.clientmessages;

public class ResourceGvk {

    private String resourceName;
    private String resourceNamespace;
    private String group;
    private String version;
    private String kind;

    public ResourceGvk(String resourceName, String resourceNamespace, String group, String version, String kind) {
        this.resourceName = resourceName;
        this.resourceNamespace = resourceNamespace;
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