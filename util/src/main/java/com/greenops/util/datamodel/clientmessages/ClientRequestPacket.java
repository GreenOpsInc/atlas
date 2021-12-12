package com.greenops.util.datamodel.clientmessages;

public class ClientRequestPacket {
    private int retryCount;
    private String namespace;
    private ClientRequest clientRequest;

    public ClientRequestPacket(String namespace, ClientRequest clientRequest) {
        this.retryCount = 0;
        this.namespace = namespace;
        this.clientRequest = clientRequest;
    }

    public ClientRequest getClientRequest() {
        return clientRequest;
    }

    public int getRetryCount() {
        return retryCount;
    }

    public void setRetryCount(int retryCount) {
        this.retryCount = retryCount;
    }

    public String getNamespace() {
        return namespace;
    }
}
