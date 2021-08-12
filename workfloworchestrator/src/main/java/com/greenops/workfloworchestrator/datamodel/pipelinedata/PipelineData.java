package com.greenops.workfloworchestrator.datamodel.pipelinedata;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;

import java.util.List;

@JsonDeserialize(as = PipelineDataImpl.class)
public interface PipelineData {

    @Deprecated
    public String getName();
    public StepData getStep(String stepName);
    public List<String> getChildrenSteps(String stepName);
    public List<String> getParentSteps(String stepName);
    public List<String> getAllSteps();
    public String getClusterName();
    public boolean isArgoVersionLock();
}
