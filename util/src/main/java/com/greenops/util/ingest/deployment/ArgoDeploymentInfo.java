package com.greenops.util.ingest.deployment;

public class ArgoDeploymentInfo {

    public static ArgoDeploymentInfo NO_OP_ARGO_DEPLOYMENT = new ArgoDeploymentInfo("", "");

    private String argoApplicationName;
    private String argoRevisionHash;

    public ArgoDeploymentInfo(String argoApplicationName, String argoRevisionHash) {
        this.argoApplicationName = argoApplicationName;
        this.argoRevisionHash = argoRevisionHash;
    }

    public String getArgoApplicationName() {
        return argoApplicationName;
    }

    public String getArgoRevisionHash() {
        return argoRevisionHash;
    }
}
