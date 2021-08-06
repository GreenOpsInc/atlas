package com.greenops.util.datamodel.clientmessages;

public class ClientDeleteByGvkRequest implements ClientRequest {

    private String orgName;
    private String type;
    private String resourceName;
    private String resourceNamespace;
    private String group;
    private String version;
    private String kind;

    public ClientDeleteByGvkRequest(String orgName, String type, String resourceName, String resourceNamespace, String group, String version, String kind) {
        this.orgName = orgName;
        this.type = type;
        this.resourceName = resourceName;
        this.resourceNamespace = resourceNamespace;
        this.group = group;
        this.version = version;
        this.kind = kind;
    }
}
