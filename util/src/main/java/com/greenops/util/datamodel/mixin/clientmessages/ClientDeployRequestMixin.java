package com.greenops.util.datamodel.mixin.clientmessages;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class ClientDeployRequestMixin {

    @JsonProperty("orgName")
    String orgName;
    @JsonProperty("deployType")
    String deployType;
    @JsonProperty("payload")
    String payload;

    @JsonCreator
    public ClientDeployRequestMixin(@JsonProperty("orgName") String orgName,
                                    @JsonProperty("deployType") String deployType,
                                    @JsonProperty("payload") String payload) {
    }
}
