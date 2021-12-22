package com.greenops.util.datamodel.mixin.clientmessages;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.greenops.util.datamodel.clientmessages.ResourcesGvkRequest;

public abstract class LabelRequestMixin {

    @JsonProperty("orgName")
    private String orgName;
    @JsonProperty("teamName")
    private String teamName;
    @JsonProperty("pipelineName")
    private String pipelineName;
    @JsonProperty("pipelineUvn")
    private String uvn;
    @JsonProperty("stepName")
    private String stepName;
    @JsonProperty(value = "clusterName")
    private String clusterName;
    @JsonProperty("resourcesGvkRequest")
    private ResourcesGvkRequest resourcesGvkRequest;
    @JsonProperty("finalTry")
    private boolean finalTry;
    @JsonProperty(value = "requestId")
    private String requestId;


    @JsonCreator
    public LabelRequestMixin(@JsonProperty("orgName") String orgName,
                             @JsonProperty("teamName") String teamName,
                             @JsonProperty("pipelineName") String pipelineName,
                             @JsonProperty("pipelineUvn") String uvn,
                             @JsonProperty("stepName") String stepName,
                             @JsonProperty(value = "clusterName") String clusterName,
                             @JsonProperty(value = "requestId") String requestId) {
    }

}
