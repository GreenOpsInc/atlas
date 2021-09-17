package com.greenops.util.datamodel.request;

public class DeployResponse {

    private final boolean success;
    private final String resourceName;
    private final String applicationNamespace;
    private final String revisionHash;

    DeployResponse(boolean success, String resourceName, String applicationNamespace, String revisionHash) {
        this.success = success;
        this.resourceName = resourceName;
        this.applicationNamespace = applicationNamespace;
        this.revisionHash = revisionHash;
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

    public String getRevisionHash() {
        return revisionHash;
    }
}
