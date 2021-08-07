package com.greenops.util.datamodel.mixin.pipelinestatus;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class FailedStepMixin {

    @JsonProperty("step")
    String step;

    @JsonProperty("deploymentFailed")
    boolean deploymentFailed;

    @JsonProperty("brokenTest")
    String brokenTest;

    @JsonProperty("brokenTestLog")
    String brokenTestLog;

    @JsonCreator
    public FailedStepMixin(@JsonProperty("step") String step,
                           @JsonProperty("deploymentFailed") boolean deploymentFailed,
                           @JsonProperty("brokenTest") String brokenTest,
                           @JsonProperty("brokenTestLog") String brokenTestLog) {
    }
}
