package com.greenops.util.datamodel.mixin.auditlog;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class DeploymentLogMixin {

    @JsonProperty(value = "pipelineUniqueVersionNumber")
    private String pipelineUniqueVersionNumber;

    @JsonProperty(value = "rollbackUniqueVersionNumber")
    private String rollbackUniqueVersionNumber;

    @JsonProperty(value = "uniqueVersionInstance")
    private int uniqueVersionInstance;

    @JsonProperty(value = "status")
    private String status;

    @JsonProperty(value = "deploymentComplete")
    private boolean deploymentComplete;

    @JsonProperty(value = "argoApplicationName")
    private String argoApplicationName;

    @JsonProperty(value = "argoRevisionHash")
    private String argoRevisionHash;

    @JsonProperty(value = "gitCommitVersion")
    private String gitCommitVersion;

    @JsonProperty(value = "brokenTest")
    private String brokenTest;

    @JsonProperty(value = "brokenTestLog")
    private String brokenTestLog;

    @JsonCreator
    DeploymentLogMixin(@JsonProperty(value = "pipelineUniqueVersionNumber") String pipelineUniqueVersionNumber,
                       @JsonProperty(value = "rollbackUniqueVersionNumber") String rollbackUniqueVersionNumber,
                       @JsonProperty(value = "uniqueVersionInstance") int uniqueVersionInstance,
                       @JsonProperty(value = "status") String status,
                       @JsonProperty(value = "deploymentComplete") boolean deploymentComplete,
                       @JsonProperty(value = "argoApplicationName") String argoApplicationName,
                       @JsonProperty(value = "argoRevisionHash") String argoRevisionHash,
                       @JsonProperty(value = "gitCommitVersion") String gitCommitVersion,
                       @JsonProperty(value = "brokenTest") String brokenTest,
                       @JsonProperty(value = "brokenTestLog") String brokenTestLog) {
    }
}
