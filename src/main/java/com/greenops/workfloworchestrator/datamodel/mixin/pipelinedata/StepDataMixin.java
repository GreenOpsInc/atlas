package com.greenops.workfloworchestrator.datamodel.mixin.pipelinedata;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.Test;

import java.util.List;

public abstract class StepDataMixin {

    @JsonProperty(value = "name")
    private String name;
    @JsonProperty(value = "argo_application")
    private String argoApplication;
    @JsonProperty(value = "application_path")
    private String argoApplicationPath;
    @JsonProperty(value = "additional_deployments")
    private String otherDeploymentsPath;
    @JsonProperty(value = "rollback")
    private boolean rollback;
    @JsonProperty(value = "tests")
    private List<Test> tests;
    @JsonProperty(value = "dependencies")
    private List<String> dependencies;

    public StepDataMixin(@JsonProperty(value = "name") String name,
                         @JsonProperty(value = "argo_application") String argoApplication,
                         @JsonProperty(value = "application_path") String argoApplicationPath,
                         @JsonProperty(value = "additional_deployments") String otherDeploymentsPath,
                         @JsonProperty(value = "rollback") boolean rollback,
                         @JsonProperty(value = "tests") List<Test> tests,
                         @JsonProperty(value = "dependencies") List<String> dependencies) {
    }

    @JsonIgnore
    abstract List<String> getName();

    @JsonIgnore
    abstract List<String> getDependencies();
}
