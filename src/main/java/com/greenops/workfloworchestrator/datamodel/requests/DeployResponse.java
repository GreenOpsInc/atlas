package com.greenops.workfloworchestrator.datamodel.requests;

public class DeployResponse {

    private final boolean success;
    private final String applicationNamespace;

    DeployResponse(boolean success, String applicationNamespace) {
        this.success = success;
        this.applicationNamespace = applicationNamespace;
    }

    public boolean getSuccess() {
        return success;
    }

    public String getApplicationNamespace() {
        return applicationNamespace;
    }
}
