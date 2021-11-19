package com.greenops.workflowtrigger.datamodel.pipelinedata;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import com.greenops.util.datamodel.event.PipelineTriggerEvent;

import java.util.List;

@JsonDeserialize(as = StepDataImpl.class)
public interface StepData {

    String getName();

    String getArgoApplicationPath();

    String getClusterName();

    void setClusterName(String clusterName);
}
