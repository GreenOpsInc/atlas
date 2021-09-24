package com.greenops.util.datamodel.clientmessages;

public class ClientDeleteByGvkRequest implements ClientRequest {

    private String orgName;
    private String teamName;
    private String pipelineName;
    private String uvn;
    private String stepName;
    private String deleteType;
    private String resourceName;
    private String resourceNamespace;
    private String group;
    private String version;
    private String kind;

    public ClientDeleteByGvkRequest(String orgName, String teamName, String pipelineName, String uvn, String stepName, String deleteType, String resourceName, String resourceNamespace, String group, String version, String kind) {
        this.orgName = orgName;
        this.teamName = teamName;
        this.pipelineName = pipelineName;
        this.uvn = uvn;
        this.stepName = stepName;
        this.deleteType = deleteType;
        this.resourceName = resourceName;
        this.resourceNamespace = resourceNamespace;
        this.group = group;
        this.version = version;
        this.kind = kind;
    }
}
