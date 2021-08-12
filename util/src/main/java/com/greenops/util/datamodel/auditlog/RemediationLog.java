package com.greenops.util.datamodel.auditlog;

import java.util.List;

public class RemediationLog implements Log {

    private final String pipelineUniqueVersionNumber;
    private final int uniqueVersionInstance;
    private List<String> unhealthyResources;
    private boolean stateRemediated;

    public RemediationLog(String pipelineUniqueVersionNumber, int uniqueVersionInstance, List<String> unhealthyResources) {
        this(pipelineUniqueVersionNumber, uniqueVersionInstance, unhealthyResources, false);
    }

    public RemediationLog(String pipelineUniqueVersionNumber, int uniqueVersionInstance, List<String> unhealthyResources, boolean stateRemediated) {
        this.pipelineUniqueVersionNumber = pipelineUniqueVersionNumber;
        this.uniqueVersionInstance = uniqueVersionInstance;
        this.unhealthyResources = unhealthyResources;
        this.stateRemediated = stateRemediated;
    }

    @Override
    public String getPipelineUniqueVersionNumber() {
        return pipelineUniqueVersionNumber;
    }

    @Override
    public int getUniqueVersionInstance() {
        return uniqueVersionInstance;
    }

    public boolean isStateRemediated() {
        return stateRemediated;
    }

    public void setStateRemediated(boolean stateRemediated) {
        this.stateRemediated = stateRemediated;
    }
}
