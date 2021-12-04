package com.greenops.util.datamodel.clientmessages;

public class ClientRequestPacket {
    private int retryCount;
    private ClientRequest clientRequest;

    public ClientRequestPacket(ClientRequest clientRequest) {
        this.retryCount = 0;
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
}
