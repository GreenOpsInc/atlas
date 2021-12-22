package com.greenops.util.datamodel.clientmessages;

public class ClientRollbackAndWatchRequest implements ClientRequest {

    //Rollback
    private String orgName;
    private String appName;
    private String revisionHash;
    //For watching
    private String watchType;
    private String uvn;
    private String teamName;
    private String pipelineName;
    private String stepName;
    private boolean finalTry;

    public ClientRollbackAndWatchRequest(String orgName, String uvn, String appName, String revisionHash, String watchType, String teamName, String pipelineName, String stepName) {
        this.orgName = orgName;
        this.uvn = uvn;
        this.appName = appName;
        this.revisionHash = revisionHash;
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
