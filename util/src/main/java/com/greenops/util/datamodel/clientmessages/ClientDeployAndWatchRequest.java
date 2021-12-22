package com.greenops.util.datamodel.clientmessages;

public class ClientDeployAndWatchRequest implements ClientRequest {

    //For deploying
    private String orgName;
    private String deployType;
    private String revisionHash;
    private String payload;
    //For watching
    private String watchType;
    private String uvn;
    private String teamName;
    private String pipelineName;
    private String stepName;
    private int testNumber;
    private boolean finalTry;

    public ClientDeployAndWatchRequest(String orgName, String uvn, String deployType, String revisionHash, String payload, String watchType, String teamName, String pipelineName, String stepName, int testNumber) {
        this.orgName = orgName;
        this.uvn = uvn;
        this.deployType = deployType;
        this.revisionHash = revisionHash;
        this.payload = payload;
        this.watchType = watchType;
        this.teamName = teamName;
        this.pipelineName = pipelineName;
        this.stepName = stepName;
        this.testNumber = testNumber;
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
