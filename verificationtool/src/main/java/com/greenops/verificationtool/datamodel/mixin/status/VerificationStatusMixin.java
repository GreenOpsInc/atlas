package com.greenops.verificationtool.datamodel.mixin.status;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.HashMap;

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
    @JsonProperty(value = "expectedDiff")
    HashMap<String, String> expectedDiff;

    @JsonCreator
    public VerificationStatusMixin(@JsonProperty(value = "status") String status,
                                   @JsonProperty(value = "eventFailed") String eventFailed,
                                   @JsonProperty(value = "stepName") String stepName,
                                   @JsonProperty(value = "failedType") String failedType,
                                   @JsonProperty(value = "log") String log,
                                   @JsonProperty(value = "expectedDiff") HashMap<String, String> expectedDiff) {

    }
}
