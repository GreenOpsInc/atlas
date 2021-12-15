package com.greenops.util.datamodel.mixin.clientmessages;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.greenops.util.datamodel.clientmessages.ResourcesGvkRequest;

public abstract class ClientSelectiveSyncAndWatchRequestMixin {

    @JsonProperty("orgName")
    String orgName;
    @JsonProperty("teamName")
    String teamName;
    @JsonProperty("pipelineName")
    String pipelineName;
    @JsonProperty("pipelineUvn")
    String uvn;
    @JsonProperty("stepName")
    String stepName;
    @JsonProperty("appName")
    String appName;
    @JsonProperty("revisionHash")
    String revisionHash;
    @JsonProperty("resourcesGvkRequest")
    ResourcesGvkRequest resourcesGvkRequest;
    @JsonProperty("finalTry")
    boolean finalTry;

    @JsonCreator
    public ClientSelectiveSyncAndWatchRequestMixin(@JsonProperty("orgName") String orgName,
                                                   @JsonProperty("teamName") String teamName,
                                                   @JsonProperty("pipelineName") String pipelineName,
                                                   @JsonProperty("pipelineUvn") String uvn,
                                                   @JsonProperty("stepName") String stepName,
                                                   @JsonProperty("appName") String appName,
                                                   @JsonProperty("revisionHash") String revisionHash,
                                                   @JsonProperty("resourcesGvkRequest") ResourcesGvkRequest resourcesGvkRequest) {
    }
}
