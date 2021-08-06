package com.greenops.util.datamodel.clientmessages;

public class ClientDeployAndWatchRequest implements ClientRequest {

    //For deploying
    private String orgName;
    private String deployType;
    private String payload;
    //For watching
    private String watchType;
    private String teamName;
    private String pipelineName;
    private String stepName;

    public ClientDeployAndWatchRequest(String orgName, String deployType, String payload, String watchType, String teamName, String pipelineName, String stepName) {
        this.orgName = orgName;
        this.deployType = deployType;
        this.payload = payload;
        this.watchType = watchType;
        this.teamName = teamName;
        this.pipelineName = pipelineName;
        this.stepName = stepName;
    }
}
