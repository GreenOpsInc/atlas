package com.greenops.workflowtrigger.datamodel.pipelinedata;

public class StepDataImpl implements StepData {
    private String name;
    private String argoApplicationPath;
    private String clusterName;

    StepDataImpl(String name,
                 String argoApplicationPath,
                 String clusterName) {
        this.name = name;
        this.argoApplicationPath = argoApplicationPath;
        this.clusterName = clusterName;
    }

    @Override
    public String getName() {
        return name;
    }

    @Override
    public String getArgoApplicationPath() {
        return argoApplicationPath;
    }

    @Override
    public String getClusterName() {
        return clusterName;
    }

    @Override
    public void setClusterName(String clusterName) {
        this.clusterName = clusterName;
    }
}
