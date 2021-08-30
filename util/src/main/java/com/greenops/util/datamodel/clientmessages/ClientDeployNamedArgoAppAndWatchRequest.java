package com.greenops.util.datamodel.clientmessages;

public class ClientDeployNamedArgoAppAndWatchRequest implements ClientRequest {

    //For deploying
    private String orgName;
    private String deployType;
    private String appName;
    //For watching
    private String watchType;
    private String teamName;
    private String pipelineName;
    private String stepName;


    public ClientDeployNamedArgoAppAndWatchRequest(String orgName, String deployType, String appName, String watchType, String teamName, String pipelineName, String stepName) {
        this.orgName = orgName;
        this.deployType = deployType;
        this.appName = appName;
        this.watchType = watchType;
        this.teamName = teamName;
        this.pipelineName = pipelineName;
        this.stepName = stepName;
    }
}
