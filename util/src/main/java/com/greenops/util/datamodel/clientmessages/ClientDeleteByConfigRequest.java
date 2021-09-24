package com.greenops.util.datamodel.clientmessages;

public class ClientDeleteByConfigRequest implements ClientRequest {

    private String orgName;
    private String teamName;
    private String pipelineName;
    private String uvn;
    private String stepName;
    private String deleteType;
    private String configPayload;

    public ClientDeleteByConfigRequest(String orgName, String teamName, String pipelineName, String uvn, String stepName, String deleteType, String configPayload) {
        this.orgName = orgName;
        this.teamName = teamName;
        this.pipelineName = pipelineName;
        this.uvn = uvn;
        this.stepName = stepName;
        this.deleteType = deleteType;
        this.configPayload = configPayload;
    }
}
