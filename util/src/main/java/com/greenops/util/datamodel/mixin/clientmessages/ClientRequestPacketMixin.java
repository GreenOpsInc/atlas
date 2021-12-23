package com.greenops.util.datamodel.mixin.clientmessages;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.greenops.util.datamodel.clientmessages.ClientRequest;

public abstract class ClientRequestPacketMixin {
    @JsonProperty("retryCount")
    int retryCount;
    @JsonProperty("namespace")
    String namespace;
    @JsonProperty("clientRequest")
    ClientRequest clientRequest;

    @JsonCreator
    public ClientRequestPacketMixin(@JsonProperty("namespace") String namespace, @JsonProperty("clientRequest") ClientRequest clientRequest) {
    }
}
