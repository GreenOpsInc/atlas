package com.greenops.util.datamodel.clientmessages;

public class ClientDeployNamedArgoApplicationRequest implements ClientRequest {

    private String orgName;
    private String deployType;
    private String appName;

    public ClientDeployNamedArgoApplicationRequest(String orgName, String deployType, String appName) {
        this.orgName = orgName;
        this.deployType = deployType;
        this.appName = appName;
    }
}
