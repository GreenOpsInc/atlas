package com.greenops.util.datamodel.pipelinestatus;

import com.greenops.util.datamodel.auditlog.DeploymentLog;

import java.util.ArrayList;
import java.util.List;

public class PipelineStatus {
    private boolean stable;
    private boolean complete;

    private List<FailedStep> failedSteps;

    public PipelineStatus() {
        this.stable = true;
        this.complete = true;
        this.failedSteps = new ArrayList<FailedStep>();
    }

    public void markIncomplete() {
        this.complete = false;
    }

    public void addDeploymentLog(DeploymentLog log, String step) {
        if (!log.isDeploymentComplete()) {
            this.complete = false;
        }

        // deployment failed
        if (log.getStatus() == DeploymentLog.DeploymentStatus.FAILURE.name()) {
            this.stable = false;
            addFailedDeploymentLog(log, step);
        }
    }

    public void addFailedDeploymentLog(DeploymentLog log, String step) {
        this.failedSteps.add(new FailedStep(step, !log.isDeploymentComplete(), log.getBrokenTest(), log.getBrokenTestLog()));
    }
}
