package com.greenops.util.datamodel.clientmessages;

public class ClientSelectiveSyncAndWatchRequest implements ClientRequest {

    private String orgName;
    private String teamName;
    private String pipelineName;
    private String stepName;
    private String appName;
    private String revisionHash;
    private ResourcesGvkRequest resourcesGvkRequest;

    public ClientSelectiveSyncAndWatchRequest(String orgName, String teamName, String pipelineName, String stepName, String appName, String revisionHash, ResourcesGvkRequest resourcesGvkRequest) {
        this.orgName = orgName;
        this.revisionHash = revisionHash;
        this.resourcesGvkRequest = resourcesGvkRequest;
        this.teamName = teamName;
        this.pipelineName = pipelineName;
        this.stepName = stepName;
        this.appName = appName;
    }
}
