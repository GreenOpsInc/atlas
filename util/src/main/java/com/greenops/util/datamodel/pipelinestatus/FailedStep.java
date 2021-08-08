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
}
