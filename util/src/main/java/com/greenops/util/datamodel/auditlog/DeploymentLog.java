package com.greenops.util.datamodel.auditlog;

import java.util.UUID;

public class DeploymentLog implements Log {

    public enum DeploymentStatus {
        SUCCESS,
        PROGRESSING,
        FAILURE
    }

    private final String pipelineUniqueVersionNumber;
    private final String rollbackUniqueVersionNumber;
    private final int uniqueVersionInstance;
    private String status;
    private boolean deploymentComplete;
    private String argoApplicationName;
    private String argoRevisionHash;
    private String gitCommitVersion;
    private String brokenTest;
    private String brokenTestLog;

    public DeploymentLog(String pipelineUniqueVersionNumber, String rollbackUniqueVersionNumber, int uniqueVersionInstance, String status, boolean deploymentComplete, String argoApplicationName, String argoRevisionHash, String gitCommitVersion, String brokenTest, String brokenTestLog) {
        this.pipelineUniqueVersionNumber = pipelineUniqueVersionNumber == null ? UUID.randomUUID().toString() : pipelineUniqueVersionNumber;
        this.rollbackUniqueVersionNumber = rollbackUniqueVersionNumber;
        this.uniqueVersionInstance = uniqueVersionInstance;
        this.status = status;
        this.argoApplicationName = argoApplicationName;
        this.deploymentComplete = deploymentComplete;
        this.argoRevisionHash = argoRevisionHash;
        this.gitCommitVersion = gitCommitVersion;
        this.brokenTest = brokenTest;
        this.brokenTestLog = brokenTestLog;
    }

    public DeploymentLog(String pipelineUniqueVersionNumber, String status, boolean deploymentComplete, String argoRevisionHash, String gitCommitVersion) {
        this(pipelineUniqueVersionNumber, null, 0, status, deploymentComplete, null, argoRevisionHash, gitCommitVersion, null, null);
    }

    @Override
    public String getPipelineUniqueVersionNumber() {
        return pipelineUniqueVersionNumber;
    }

    public String getRollbackUniqueVersionNumber() {
        return rollbackUniqueVersionNumber;
    }

    @Override
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

    public String getArgoRevisionHash() {
        return argoRevisionHash;
    }

    public void setArgoRevisionHash(String argoRevisionHash) {
        this.argoRevisionHash = argoRevisionHash;
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
