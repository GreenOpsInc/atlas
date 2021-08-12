package com.greenops.workfloworchestrator.datamodel.mixin.pipelinedata;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.StepData;

import java.util.List;
import java.util.Map;

public abstract class PipelineDataMixin {

    @JsonProperty(value = "name")
    String name;
    @JsonProperty(value = "cluster_name")
    String clusterName;
    @JsonProperty(value = "argo_version_lock")
    boolean argoVersionLock;
    @JsonProperty(value = "steps")
    List<StepData> steps;
    @JsonIgnore
    Map<String, List<String>> stepParents;
    @JsonIgnore
    Map<String, List<String>> stepChildren;

    @JsonCreator
    public PipelineDataMixin(@JsonProperty(value = "name") String name,
                             @JsonProperty(value = "cluster_name") String clusterName,
                             @JsonProperty(value = "argo_version_lock") boolean argoVersionLock,
                             @JsonProperty(value = "steps") List<StepData> stepDataList) {
    }

    @JsonIgnore
    abstract List<String> getAllSteps();

}
