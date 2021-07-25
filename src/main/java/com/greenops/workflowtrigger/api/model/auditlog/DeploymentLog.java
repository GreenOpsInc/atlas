package com.greenops.workflowtrigger.api.model.auditlog;

import java.util.UUID;

public class DeploymentLog {

    public enum DeploymentStatus {
        SUCCESS,
        PROGRESSING,
        FAILURE
    }

    private final String uniqueVersionNumber;
    private final int uniqueVersionInstance;
    private String status;
    private boolean deploymentComplete;
    private String argoApplicationName;
    private int argoRevisionId;
    private String gitCommitVersion;
    private String brokenTest;
    private String brokenTestLog;

    public DeploymentLog(String uniqueVersionNumber, int uniqueVersionInstance, String status, boolean deploymentComplete, String argoApplicationName, int argoRevisionId, String gitCommitVersion, String brokenTest, String brokenTestLog) {
        this.uniqueVersionNumber = uniqueVersionNumber;
        this.uniqueVersionInstance = uniqueVersionInstance;
        this.status = status;
        this.argoApplicationName = argoApplicationName;
        this.deploymentComplete = deploymentComplete;
        this.argoRevisionId = argoRevisionId;
        this.gitCommitVersion = gitCommitVersion;
        this.brokenTest = brokenTest;
        this.brokenTestLog = brokenTestLog;
    }

    public DeploymentLog(String status, boolean deploymentComplete, int argoRevisionId, String gitCommitVersion) {
        this(UUID.randomUUID().toString(), 0, status, deploymentComplete, null, argoRevisionId, gitCommitVersion, null, null);
    }

    public String getUniqueVersionNumber() {
        return uniqueVersionNumber;
    }

    public int getUniqueVersionInstance() {
        return uniqueVersionInstance;
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

    public String getArgoApplicationName() {
        return argoApplicationName;
    }

    public void setArgoApplicationName(String argoApplicationName) {
        this.argoApplicationName = argoApplicationName;
    }

    public int getArgoRevisionId() {
        return argoRevisionId;
    }

    public void setArgoRevisionId(int argoRevisionId) {
        this.argoRevisionId = argoRevisionId;
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
