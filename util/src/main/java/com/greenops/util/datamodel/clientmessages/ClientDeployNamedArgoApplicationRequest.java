package com.greenops.util.datamodel.clientmessages;

public class ClientDeployNamedArgoApplicationRequest implements ClientRequest {

    private String orgName;
    private String deployType;
    private String appName;
    private boolean finalTry;

    public ClientDeployNamedArgoApplicationRequest(String orgName, String deployType, String appName) {
        this.orgName = orgName;
        this.deployType = deployType;
        this.appName = appName;
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
