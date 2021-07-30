package com.greenops.workfloworchestrator.datamodel.requests;

public class DeployResponse {

    private final boolean success;
    private final String resourceName;
    private final String applicationNamespace;

    DeployResponse(boolean success, String resourceName, String applicationNamespace) {
        this.success = success;
        this.resourceName = resourceName;
        this.applicationNamespace = applicationNamespace;
    }

    public boolean getSuccess() {
        return success;
    }

    public String getResourceName() {
        return resourceName;
    }

    public String getApplicationNamespace() {
        return applicationNamespace;
    }
}
