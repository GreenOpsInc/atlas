package com.greenops.util.datamodel.mixin.auditlog;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.List;

public abstract class RemediationLogMixin {

    @JsonProperty(value = "pipelineUniqueVersionNumber")
    private String pipelineUniqueVersionNumber;

    @JsonProperty(value = "uniqueVersionInstance")
    private int uniqueVersionInstance;

    @JsonProperty(value = "unhealthyResources")
    private List<String> unhealthyResources;

    @JsonProperty(value = "remediationStatus")
    private String remediationStatus;

    @JsonCreator
    RemediationLogMixin(@JsonProperty(value = "pipelineUniqueVersionNumber") String pipelineUniqueVersionNumber,
                        @JsonProperty(value = "uniqueVersionInstance") int uniqueVersionInstance,
                        @JsonProperty(value = "unhealthyResources") List<String> unhealthyResources,
                        @JsonProperty(value = "remediationStatus") String remediationStatus) {
    }
}
