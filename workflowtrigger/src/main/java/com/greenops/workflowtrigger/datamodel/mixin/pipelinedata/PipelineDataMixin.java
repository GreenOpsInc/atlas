package com.greenops.workflowtrigger.datamodel.mixin.pipelinedata;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.greenops.workflowtrigger.datamodel.pipelinedata.StepData;

import java.util.List;

public abstract class PipelineDataMixin {

    @JsonProperty(value = "cluster_name")
    String clusterName;
    @JsonProperty(value = "steps")
    List<StepData> steps;

    @JsonCreator
    public PipelineDataMixin(@JsonProperty(value = "cluster_name") String clusterName,
                             @JsonProperty(value = "steps") List<StepData> stepDataList) {
    }

    @JsonIgnore
    abstract List<String> getAllSteps();

}
