package com.greenops.workfloworchestrator.datamodel.pipelinedata;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;

import java.util.List;

@JsonDeserialize(as = StepDataImpl.class)
public interface StepData {
    String getName();
    List<String> getDependencies();

    static StepData createRootStep() {
        return new StepDataImpl("ATLAS_ROOT_DATA", null, null, null, false, List.of(), List.of());
    }
}
