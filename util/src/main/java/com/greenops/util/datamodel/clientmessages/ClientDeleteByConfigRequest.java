package com.greenops.util.datamodel.clientmessages;

public class ClientDeleteByConfigRequest implements ClientRequest {

    private String orgName;
    private String type;
    private String configPayload;

    public ClientDeleteByConfigRequest(String orgName, String type, String configPayload) {
        this.orgName = orgName;
        this.type = type;
        this.configPayload = configPayload;
    }
}
