package com.greenops.util.datamodel.clientmessages;

public class LabelRequest implements ClientRequest, NotificationRequest {

    private String orgName;
    private String teamName;
    private String pipelineName;
    private String uvn;
    private String stepName;
    private String clusterName;
    private ResourcesGvkRequest resourcesGvkRequest;
    private boolean finalTry;
    private String requestId;


    public LabelRequest(String orgName, String teamName, String pipelineName, String uvn, String stepName, String clusterName, ResourcesGvkRequest resourcesGvkRequest, String requestId) {
        this.orgName = orgName;
        this.uvn = uvn;
        this.resourcesGvkRequest = resourcesGvkRequest;
        this.teamName = teamName;
        this.pipelineName = pipelineName;
        this.stepName = stepName;
        this.clusterName = clusterName;
        this.requestId = requestId;
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

    @Override
    public void setRequestId(String requestId) {
        this.requestId = requestId;
    }
}
