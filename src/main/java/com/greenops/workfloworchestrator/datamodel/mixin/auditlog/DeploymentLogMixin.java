package com.greenops.workfloworchestrator.datamodel.mixin.auditlog;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

public class DeploymentLogMixin {

    enum Status {
        SUCCESS,
        PROGRESSING,
        FAILURE
    }

    @JsonProperty(value = "status")
    private String status;

    @JsonProperty(value = "deploymentComplete")
    private boolean deploymentComplete;

    @JsonProperty(value = "gitCommitVersion")
    private String gitCommitVersion;

    @JsonProperty(value = "brokenTest")
    private String brokenTest;

    @JsonProperty(value = "brokenTestLog")
    private String brokenTestLog;

    @JsonCreator
    DeploymentLogMixin(@JsonProperty(value = "status") String status,
                       @JsonProperty(value = "deploymentComplete") boolean deploymentComplete,
                       @JsonProperty(value = "gitCommitVersion") String gitCommitVersion,
                       @JsonProperty(value = "brokenTest") String brokenTest,
                       @JsonProperty(value = "brokenTestLog") String brokenTestLog) {
    }
}
