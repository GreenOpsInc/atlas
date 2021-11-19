package com.greenops.workflowtrigger.datamodel.pipelinedata;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;

import java.util.List;

@JsonDeserialize(as = PipelineDataImpl.class)
public interface PipelineData {

    public StepData getStep(String stepName);
    public List<String> getAllSteps();
    public String getClusterName();
}
