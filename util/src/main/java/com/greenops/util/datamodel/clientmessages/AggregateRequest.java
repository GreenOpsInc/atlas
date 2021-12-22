package com.greenops.util.datamodel.clientmessages;

public class AggregateRequest implements ClientRequest, NotificationRequest {

    private String orgName;
    private String teamName;
    private String pipelineName;
    private String uvn;
    private String stepName;
    private String clusterName;
    private	String namespace;
    private boolean finalTry;
    private String requestId;


    public AggregateRequest(String orgName, String teamName, String pipelineName, String uvn, String stepName, String clusterName, String namespace, String requestId) {
        this.orgName = orgName;
        this.uvn = uvn;
        this.teamName = teamName;
        this.pipelineName = pipelineName;
        this.stepName = stepName;
        this.clusterName = clusterName;
        this.namespace = namespace;
        this.requestId = requestId;
        this.finalTry = false;
    }

    public String getClusterName() {
        return clusterName;
    }

    public String getNamespace() {
        return namespace;
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
