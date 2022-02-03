package com.greenops.util.datamodel.pipelinestatus;

public class FailedStep {
    private String step;
    private boolean deploymentFailed;
    private String brokenTest;
    private String brokenTestLog;

    public FailedStep(String step, boolean deploymentFailed, String brokenTest, String brokenTestLog) {
        this.step = step;
        this.deploymentFailed = deploymentFailed;
        this.brokenTest = brokenTest;
        this.brokenTestLog = brokenTestLog;
    }

    public String getStep() {
        return step;
    }

    public boolean isDeploymentFailed() {
        return deploymentFailed;
    }

    public String getBrokenTest() {
        return brokenTest;
    }

    public String getBrokenTestLog() {
        return brokenTestLog;
    }
}
