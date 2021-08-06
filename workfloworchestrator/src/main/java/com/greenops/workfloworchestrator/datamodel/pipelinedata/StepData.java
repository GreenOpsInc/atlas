package com.greenops.workfloworchestrator.datamodel.pipelinedata;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;

import java.util.List;

@JsonDeserialize(as = StepDataImpl.class)
public interface StepData {
    static final String ROOT_STEP_NAME = "ATLAS_ROOT_DATA";

    String getName();

    String getArgoApplication();

    String getArgoApplicationPath();

    String getOtherDeploymentsPath();

    String getClusterName();

    void setClusterName(String clusterName);

    boolean getRollback();

    List<Test> getTests();

    List<String> getDependencies();

    static StepData createRootStep() {
        return new StepDataImpl(ROOT_STEP_NAME, null, null, null, null, false, List.of(), List.of());
    }
}
