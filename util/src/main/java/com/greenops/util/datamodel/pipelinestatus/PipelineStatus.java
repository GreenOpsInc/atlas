package com.greenops.util.datamodel.pipelinestatus;

import com.greenops.util.datamodel.auditlog.DeploymentLog;
import com.greenops.util.datamodel.auditlog.Log;

import java.util.ArrayList;
import java.util.List;

public class PipelineStatus {

    private final List<String> progressingSteps;
    private final List<FailedStep> failedSteps;
    private boolean stable;
    private boolean complete;
    private boolean cancelled;

    public PipelineStatus() {
        this.progressingSteps = new ArrayList<>();
        this.stable = true;
        this.complete = true;
        this.cancelled = false;
        this.failedSteps = new ArrayList<>();
    }

    public List<String> getProgressingSteps() {
        return this.progressingSteps;
    }

    public boolean isStable() {
        return this.stable;
    }

    public boolean isComplete() {
        return this.complete;
    }

    public boolean isCancelled() {
        return this.cancelled;
    }

    public List<FailedStep> getFailedSteps() {
        return this.failedSteps;
    }

    public void markCancelled() {
        this.cancelled = true;
        this.complete = false;
    }

    public void markIncomplete() {
        this.complete = false;
    }

    public void addProgressingStep(String stepName) {
        progressingSteps.add(stepName);
    }

    //This method asserts the stability of the pipeline by seeing if the latest audit log shows the state to be stable
    public void addLatestLog(Log log) {
        if (log instanceof DeploymentLog) {
            var deploymentLog = (DeploymentLog) log;

            // deployment failed
            if (deploymentLog.getStatus().equals(Log.LogStatus.FAILURE.name())) {
                this.stable = false;
            }
        } else {
            //Was instance of a RemediationLog
            if (!log.getStatus().equals(Log.LogStatus.FAILURE.name())) {
                this.stable = false;
            }
            //TODO: Needs to be added to failed step
        }
    }

    public void addLatestDeploymentLog(DeploymentLog log, String step) {
        if (log.getStatus().equals(Log.LogStatus.FAILURE.name())) {
            this.complete = false;
            addFailedDeploymentLog(log, step);
        }
    }

    public void addFailedDeploymentLog(DeploymentLog log, String step) {
        this.failedSteps.add(new FailedStep(step, !log.isDeploymentComplete(), log.getBrokenTest(), log.getBrokenTestLog()));
    }
}
