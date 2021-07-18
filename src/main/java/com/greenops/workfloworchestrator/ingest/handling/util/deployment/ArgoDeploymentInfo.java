package com.greenops.workfloworchestrator.ingest.handling.util.deployment;

public class ArgoDeploymentInfo {

    public static ArgoDeploymentInfo NO_OP_ARGO_DEPLOYMENT = new ArgoDeploymentInfo("", -1);

    private String argoApplicationName;
    private int argoRevisionId;

    public ArgoDeploymentInfo(String argoApplicationName, int argoRevisionId) {
        this.argoApplicationName = argoApplicationName;
        this.argoRevisionId = argoRevisionId;
    }

    public String getArgoApplicationName() {
        return argoApplicationName;
    }

    public int getArgoRevisionId() {
        return argoRevisionId;
    }
}
