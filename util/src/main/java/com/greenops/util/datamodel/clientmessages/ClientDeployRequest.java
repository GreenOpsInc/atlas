package com.greenops.util.datamodel.clientmessages;

public class ClientDeployRequest implements ClientRequest {

    private String orgName;
    private String deployType;
    private String payload;

    public ClientDeployRequest(String orgName, String deployType, String payload) {
        this.orgName = orgName;
        this.deployType = deployType;
        this.payload = payload;
    }
}
