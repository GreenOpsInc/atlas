package com.greenops.util.datamodel.clientmessages;

public class ClientDeployRequest implements ClientRequest {

    private String orgName;
    private String teamName;
    private String pipelineName;
    private String uvn;
    private String stepName;
    private String responseEventType;
    private String deployType;
    private String revisionHash;
    private String payload;

    public ClientDeployRequest(String orgName, String teamName, String pipelineName, String uvn, String stepName, String responseEventType, String deployType, String revisionHash, String payload) {
        this.orgName = orgName;
        this.teamName = teamName;
        this.pipelineName = pipelineName;
        this.uvn = uvn;
        this.stepName = stepName;
        this.responseEventType = responseEventType;
        this.deployType = deployType;
        this.revisionHash = revisionHash;
        this.payload = payload;
    }
}
