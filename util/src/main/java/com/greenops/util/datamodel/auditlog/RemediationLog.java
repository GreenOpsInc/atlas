package com.greenops.util.datamodel.auditlog;

import java.util.List;

public class RemediationLog implements Log {

    public enum RemediationStatus {
        SUCCESS,
        PROGRESSING,
        FAILURE
    }

    private final String pipelineUniqueVersionNumber;
    private final int uniqueVersionInstance;
    private List<String> unhealthyResources;
    private String remediationStatus;

    public RemediationLog(String pipelineUniqueVersionNumber, int uniqueVersionInstance, List<String> unhealthyResources) {
        this(pipelineUniqueVersionNumber, uniqueVersionInstance, unhealthyResources, RemediationStatus.PROGRESSING.name());
    }

    public RemediationLog(String pipelineUniqueVersionNumber, int uniqueVersionInstance, List<String> unhealthyResources, String remediationStatus) {
        this.pipelineUniqueVersionNumber = pipelineUniqueVersionNumber;
        this.uniqueVersionInstance = uniqueVersionInstance;
        this.unhealthyResources = unhealthyResources;
        this.remediationStatus = remediationStatus;
    }

    @Override
    public String getPipelineUniqueVersionNumber() {
        return pipelineUniqueVersionNumber;
    }

    @Override
    public int getUniqueVersionInstance() {
        return uniqueVersionInstance;
    }

    public String getRemediationStatus() {
        return remediationStatus;
    }

    public void setStateRemediated(String remediationStatus) {
        this.remediationStatus = remediationStatus;
    }
}
