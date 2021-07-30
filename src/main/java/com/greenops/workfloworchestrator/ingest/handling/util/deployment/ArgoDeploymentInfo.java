package com.greenops.workfloworchestrator.ingest.handling.util.deployment;

public class ArgoDeploymentInfo {

    public static ArgoDeploymentInfo NO_OP_ARGO_DEPLOYMENT = new ArgoDeploymentInfo("");

    private String argoApplicationName;
    private int argoRevisionId;

    public ArgoDeploymentInfo(String argoApplicationName) {
        this.argoApplicationName = argoApplicationName;
        this.argoRevisionId = -1;
    }

    public String getArgoApplicationName() {
        return argoApplicationName;
    }

    public int getArgoRevisionId() {
        return argoRevisionId;
    }
}
