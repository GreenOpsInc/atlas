package com.greenops.util.datamodel.mixin.clientmessages;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class ClientDeleteByConfigRequestMixin {

    @JsonProperty("orgName")
    String orgName;
    @JsonProperty("teamName")
    String teamName;
    @JsonProperty("pipelineName")
    String pipelineName;
    @JsonProperty("stepName")
    String stepName;
    @JsonProperty("type")
    String type;
    @JsonProperty("configPayload")
    String configPayload;

    @JsonCreator
    public ClientDeleteByConfigRequestMixin(@JsonProperty("orgName") String orgName,
                                            @JsonProperty("teamName") String teamName,
                                            @JsonProperty("pipelineName") String pipelineName,
                                            @JsonProperty("stepName") String stepName,
                                            @JsonProperty("type") String type,
                                            @JsonProperty("configPayload") String configPayload) {
    }
}
