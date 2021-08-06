package com.greenops.util.datamodel.mixin.clientmessages;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class ClientDeployNamedArgoApplicationRequestMixin {

    @JsonProperty("orgName")
    String orgName;
    @JsonProperty("type")
    String type;
    @JsonProperty("appName")
    String appName;

    @JsonCreator
    public ClientDeployNamedArgoApplicationRequestMixin(@JsonProperty("orgName") String orgName,
                                                        @JsonProperty("type") String type,
                                                        @JsonProperty("appName") String appName) {
    }
}
