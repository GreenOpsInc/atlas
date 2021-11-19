package com.greenops.workflowtrigger.datamodel.mixin.pipelinedata;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.List;

public abstract class StepDataMixin {

    @JsonProperty(value = "name")
    private String name;
    @JsonProperty(value = "application_path")
    private String argoApplicationPath;
    @JsonProperty(value = "cluster_name")
    private String clusterName;

    public StepDataMixin(@JsonProperty(value = "name") String name,
                         @JsonProperty(value = "application_path") String argoApplicationPath,
                         @JsonProperty(value = "cluster_name") String clusterName) {
    }

    @JsonIgnore
    abstract List<String> getName();
}
