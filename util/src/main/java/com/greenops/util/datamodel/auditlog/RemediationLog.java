package com.greenops.util.datamodel.auditlog;

import java.util.List;

public class RemediationLog implements Log {

    private final String pipelineUniqueVersionNumber;
    private final int uniqueVersionInstance;
    private List<String> unhealthyResources;
    private String remediationStatus;

    public RemediationLog(String pipelineUniqueVersionNumber, int uniqueVersionInstance, List<String> unhealthyResources) {
        this(pipelineUniqueVersionNumber, uniqueVersionInstance, unhealthyResources, Log.LogStatus.PROGRESSING.name());
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

    @Override
    public String getStatus() {
        return remediationStatus;
    }

    @Override
    public void setStatus(String remediationStatus) {
        this.remediationStatus = remediationStatus;
    }
}
