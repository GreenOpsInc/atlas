package com.greenops.util.datamodel.mixin.pipelinedata;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.greenops.util.datamodel.pipelinedata.Test;

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
    @JsonProperty(value = "cluster_name")
    private String clusterName;
    @JsonProperty(value = "tests")
    private List<Test> tests;
    @JsonProperty(value = "remediation_limit")
    private int remediationLimit;
    @JsonProperty(value = "rollback_limit")
    private int rollbackLimit;
    @JsonProperty(value = "dependencies")
    private List<String> dependencies;

    public StepDataMixin(@JsonProperty(value = "name") String name,
                         @JsonProperty(value = "argo_application") String argoApplication,
                         @JsonProperty(value = "application_path") String argoApplicationPath,
                         @JsonProperty(value = "additional_deployments") String otherDeploymentsPath,
                         @JsonProperty(value = "cluster_name") String clusterName,
                         @JsonProperty(value = "tests") List<Test> tests,
                         @JsonProperty(value = "remediation_limit") int remediationLimit,
                         @JsonProperty(value = "rollback_limit") int rollbackLimit,
                         @JsonProperty(value = "dependencies") List<String> dependencies) {
    }

    @JsonIgnore
    abstract List<String> getName();

    @JsonIgnore
    abstract List<String> getDependencies();
}
