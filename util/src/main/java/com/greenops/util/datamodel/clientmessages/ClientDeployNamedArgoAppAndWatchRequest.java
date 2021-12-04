package com.greenops.util.datamodel.clientmessages;

public class ClientDeployNamedArgoAppAndWatchRequest implements ClientRequest {

    //For deploying
    private String orgName;
    private String deployType;
    private String appName;
    //For watching
    private String watchType;
    private String uvn;
    private String teamName;
    private String pipelineName;
    private String stepName;
    private boolean finalTry;

    public ClientDeployNamedArgoAppAndWatchRequest(String orgName, String uvn, String deployType, String appName, String watchType, String teamName, String pipelineName, String stepName) {
        this.orgName = orgName;
        this.uvn = uvn;
        this.deployType = deployType;
        this.appName = appName;
        this.watchType = watchType;
        this.teamName = teamName;
        this.pipelineName = pipelineName;
        this.stepName = stepName;
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
