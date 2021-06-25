package com.greenops.workfloworchestrator.datamodel.pipelinedata;

import java.util.ArrayList;
import java.util.List;

public class StepDataImpl implements StepData {
    private String name;
    private String argoApplication;
    private String argoApplicationPath;
    private String otherDeploymentsPath;
    private boolean rollback;
    private List<Test> tests;
    private List<String> dependencies;

    StepDataImpl(String name,
                 String argoApplication,
                 String argoApplicationPath,
                 String otherDeploymentsPath,
                 boolean rollback,
                 List<Test> tests,
                 List<String> dependencies) {
        this.name = name;
        this.argoApplication = argoApplication;
        this.argoApplicationPath = argoApplicationPath;
        this.otherDeploymentsPath = otherDeploymentsPath;
        this.rollback = rollback;
        this.tests = tests == null ? new ArrayList<>() : tests;
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
    public boolean getRollback() {
        return rollback;
    }

    @Override
    public List<Test> getTests() {
        return tests;
    }

    @Override
    public List<String> getDependencies() {
        return dependencies;
    }
}
