package com.greenops.util.datamodel.clientmessages;

public class ClientDeployNamedArgoApplicationRequest implements ClientRequest {

    private String orgName;
    private String type;
    private String appName;

    public ClientDeployNamedArgoApplicationRequest(String orgName, String type, String appName) {
        this.orgName = orgName;
        this.type = type;
        this.appName = appName;
    }
}
