package com.greenops.workfloworchestrator.datamodel.requests;

public class DeployResponse {

    private final boolean success;
    private final String resourceName;
    private final String applicationNamespace;
    private final int revisionId;

    DeployResponse(boolean success, String resourceName, String applicationNamespace, int revisionId) {
        this.success = success;
        this.resourceName = resourceName;
        this.applicationNamespace = applicationNamespace;
        this.revisionId = revisionId;
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

    public int getRevisionId() {
        return revisionId;
    }
}
