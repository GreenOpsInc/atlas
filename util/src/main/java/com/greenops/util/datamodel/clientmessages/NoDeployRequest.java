package com.greenops.util.datamodel.clientmessages;

public class NoDeployRequest implements ClientRequest, NotificationRequest {

    private String orgName;
    private String teamName;
    private String pipelineName;
    private String uvn;
    private String stepName;
    private final String clusterName;
    private final String namespace;
    private final boolean apply;
    private String requestId;
    private boolean finalTry;

    public NoDeployRequest(String orgName, String teamName, String pipelineName, String uvn, String stepName, String clusterName, String namespace, boolean apply, String requestId) {
        this.orgName = orgName;
        this.teamName = teamName;
        this.pipelineName = pipelineName;
        this.uvn = uvn;
        this.stepName = stepName;
        this.clusterName = clusterName;
        this.namespace = namespace;
        this.apply = apply;
        this.requestId = requestId;
        this.finalTry = false;
    }

    public String getClusterName() {
        return clusterName;
    }

    public String getNamespace() {
        return namespace;
    }

    public boolean getApply() {
        return apply;
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
