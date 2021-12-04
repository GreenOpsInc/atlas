package com.greenops.util.datamodel.clientmessages;

public class ClientSelectiveSyncAndWatchRequest implements ClientRequest {

    private String orgName;
    private String uvn;
    private String teamName;
    private String pipelineName;
    private String stepName;
    private String appName;
    private String revisionHash;
    private ResourcesGvkRequest resourcesGvkRequest;
    private boolean finalTry;

    public ClientSelectiveSyncAndWatchRequest(String orgName, String teamName, String pipelineName, String uvn, String stepName, String appName, String revisionHash, ResourcesGvkRequest resourcesGvkRequest) {
        this.orgName = orgName;
        this.uvn = uvn;
        this.revisionHash = revisionHash;
        this.resourcesGvkRequest = resourcesGvkRequest;
        this.teamName = teamName;
        this.pipelineName = pipelineName;
        this.stepName = stepName;
        this.appName = appName;
        this.finalTry = false;
    }

    @Override
    public void setFinalTry(boolean finalTry) {
        this.finalTry = finalTry;
    }

    @Override
    public boolean isFinalTry() {
        return finalTry;
    }
}
