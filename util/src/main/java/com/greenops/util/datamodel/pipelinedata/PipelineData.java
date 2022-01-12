package com.greenops.util.datamodel.pipelinedata;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;

import java.util.List;

@JsonDeserialize(as = PipelineDataImpl.class)
public interface PipelineData {

    @Deprecated
    String getName();

    StepData getStep(String stepName);

    List<String> getChildrenSteps(String stepName);

    List<String> getParentSteps(String stepName);

    List<String> getAllSteps();

    List<String> getAllStepsOrdered();

    String getClusterName();

    boolean isArgoVersionLock();
}
