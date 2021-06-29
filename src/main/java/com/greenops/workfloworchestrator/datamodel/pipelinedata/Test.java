package com.greenops.workfloworchestrator.datamodel.pipelinedata;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;

import java.util.Map;

@JsonDeserialize(as = CustomTest.class)
public interface Test {
    public String getPath();
    public boolean shouldExecuteInPod();
    public boolean shouldExecuteBefore();
    public Map<String, String> getVariables();
}
