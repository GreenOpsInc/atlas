package com.greenops.workfloworchestrator.datamodel.auditlog;

public class DeploymentLog {

    public enum DeploymentStatus {
        SUCCESS,
        PROGRESSING,
        FAILURE
    }

    //TODO: Should have "progressing" if step is not complete
    private String status;
    private boolean deploymentComplete;
    private String gitCommitVersion;
    //TODO: Should have "progressing" if tests are not done yet
    private String brokenTest;
    private String brokenTestLog;

    public DeploymentLog(String status, boolean deploymentComplete, String gitCommitVersion, String brokenTest, String brokenTestLog) {
        this.status = status;
        this.deploymentComplete = deploymentComplete;
        this.gitCommitVersion = gitCommitVersion;
        this.brokenTest = brokenTest;
        this.brokenTestLog = brokenTestLog;
    }

    public DeploymentLog(String status, boolean deploymentComplete, String gitCommitVersion) {
        this(status, deploymentComplete, gitCommitVersion, null, null);
    }

    public String getStatus() {
        return status;
    }

    public void setStatus(String status) {
        this.status = status;
    }

    public boolean isDeploymentComplete() {
        return deploymentComplete;
    }

    public void setDeploymentComplete(boolean deploymentComplete) {
        this.deploymentComplete = deploymentComplete;
    }

    public String getGitCommitVersion() {
        return gitCommitVersion;
    }

    public void setGitCommitVersion(String gitCommitVersion) {
        this.gitCommitVersion = gitCommitVersion;
    }

    public String getBrokenTest() {
        return brokenTest;
    }

    public void setBrokenTest(String brokenTest) {
        this.brokenTest = brokenTest;
    }

    public String getBrokenTestLog() {
        return brokenTestLog;
    }

    public void setBrokenTestLog(String brokenTestLog) {
        this.brokenTestLog = brokenTestLog;
    }
}
