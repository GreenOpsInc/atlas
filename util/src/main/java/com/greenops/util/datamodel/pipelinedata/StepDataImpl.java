package com.greenops.util.datamodel.pipelinedata;

import java.util.ArrayList;
import java.util.List;

public class StepDataImpl implements StepData {
    private String name;
    private String argoApplication;
    private String argoApplicationPath;
    private String otherDeploymentsPath;
    private String clusterName;
    private List<Test> tests;
    private int remediationLimit;
    private int rollbackLimit;
    private List<String> dependencies;

    public StepDataImpl(String name,
                 String argoApplication,
                 String argoApplicationPath,
                 String otherDeploymentsPath,
                 String clusterName,
                 List<Test> tests,
                 int remediationLimit,
                 int rollbackLimit,
                 List<String> dependencies) {
        this.name = name;
        this.argoApplication = argoApplication;
        this.argoApplicationPath = argoApplicationPath;
        this.otherDeploymentsPath = otherDeploymentsPath;
        this.clusterName = clusterName;
        this.tests = tests == null ? new ArrayList<>() : tests;
        this.remediationLimit = remediationLimit;
        this.rollbackLimit = rollbackLimit;
        this.dependencies = dependencies == null ? new ArrayList<>() : dependencies;
    }

    @Override
    public String getName() {
        return name;
    }

    @Override
    public String getArgoApplication() {
        return argoApplication;
    }

    @Override
    public String getArgoApplicationPath() {
        return argoApplicationPath;
    }

    @Override
    public String getOtherDeploymentsPath() {
        return otherDeploymentsPath;
    }

    @Override
    public String getClusterName() {
        return clusterName;
    }

    @Override
    public void setClusterName(String clusterName) {
        this.clusterName = clusterName;
    }

    @Override
    public List<Test> getTests() {
        return tests;
    }

    @Override
    public List<String> getDependencies() {
        return dependencies;
    }

    @Override
    public int getRemediationLimit() {
        return remediationLimit;
    }

    @Override
    public int getRollbackLimit() {
        return rollbackLimit;
    }
}
