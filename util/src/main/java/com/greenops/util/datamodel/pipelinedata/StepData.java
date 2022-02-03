package com.greenops.util.datamodel.pipelinedata;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import com.greenops.util.datamodel.event.PipelineTriggerEvent;

import java.util.List;

@JsonDeserialize(as = StepDataImpl.class)
public interface StepData {
    static final String ROOT_STEP_NAME = PipelineTriggerEvent.ROOT_STEP_NAME;

    String getName();

    String getArgoApplication();

    String getArgoApplicationPath();

    String getOtherDeploymentsPath();

    String getClusterName();

    void setClusterName(String clusterName);

    List<Test> getTests();

    List<String> getDependencies();

    int getRemediationLimit();

    int getRollbackLimit();

    static StepData createRootStep() {
        return new StepDataImpl(ROOT_STEP_NAME, null, null, null, null, List.of(), 0, 0, List.of());
    }
}
