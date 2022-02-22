package com.greenops.verificationtool.datamodel.mixin.status;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class VerificationStatusMixin {
    @JsonProperty(value = "status")
    String status;
    @JsonProperty(value = "eventFailed")
    String eventFailed;
    @JsonProperty(value = "stepName")
    String stepName;
    @JsonProperty(value = "failedType")
    String failedType;
    @JsonProperty(value = "log")
    String log;

    @JsonCreator
    public VerificationStatusMixin(@JsonProperty(value = "status") String status,
                                   @JsonProperty(value = "eventFailed") String eventFailed,
                                   @JsonProperty(value = "stepName") String stepName,
                                   @JsonProperty(value = "failedType") String failedType,
                                   @JsonProperty(value = "log") String log) {

    }
}
