package com.greenops.util.datamodel.mixin.event;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class EventMixin {

    @JsonProperty(value = "orgName")
    String orgName;
    @JsonProperty(value = "teamName")
    String teamName;
    @JsonProperty(value = "pipelineName")
    String pipelineName;
    @JsonProperty(value = "pipelineUvn")
    String pipelineUvn;
    @JsonProperty(value = "stepName")
    String stepName;

    //Properties to ignore
    @JsonIgnore
    @JsonProperty(value = "deliveryAttempt")
    int deliveryAttempt;

    @JsonIgnore
    abstract String getPipelineUvn();

    @JsonIgnore
    abstract String getMQKey();

    @JsonIgnore
    abstract void setDeliveryAttempt(int attempt);

    @JsonIgnore
    abstract int getDeliveryAttempt();

}
