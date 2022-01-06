package com.greenops.util.tslmanager;

public enum ClientSecretName {
    NOT_VALID_SECRET_NAME("not-valid"),
    WORKFLOW_TRIGGER_SECRET_NAME("workflowtrigger-tls"),
    CLIENT_WRAPPER_SECRET_NAME("clientwrapper-tls"),
    REPO_SERVER_SECRET_NAME("pipelinereposerver-tls"),
    COMMAND_DELEGATOR_SECRET_NAME("commanddelegator-tls"),
    ARGOCD_REPO_SERVER_SECRET_NAME("argocd-repo-server-tls"),
    KAFKA_SECRET_NAME("kafka-tls");

    private final String text;

    ClientSecretName(final String text) {
        this.text = text;
    }

    @Override
    public String toString() {
        return text;
    }
}
