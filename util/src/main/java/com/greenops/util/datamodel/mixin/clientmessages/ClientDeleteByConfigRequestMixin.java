package com.greenops.util.datamodel.mixin.clientmessages;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class ClientDeleteByConfigRequestMixin {

    @JsonProperty("orgName")
    String orgName;
    @JsonProperty("type")
    String type;
    @JsonProperty("configPayload")
    String configPayload;

    @JsonCreator
    public ClientDeleteByConfigRequestMixin(@JsonProperty("orgName") String orgName,
                                            @JsonProperty("type") String type,
                                            @JsonProperty("configPayload") String configPayload) {
    }
}
