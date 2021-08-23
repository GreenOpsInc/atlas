package com.greenops.util.datamodel.pipelinestatus;

import com.greenops.util.datamodel.auditlog.DeploymentLog;
import com.greenops.util.datamodel.auditlog.Log;
import com.greenops.util.datamodel.auditlog.RemediationLog;

import java.util.ArrayList;
import java.util.List;

public class PipelineStatus {

    private List<String> progressingSteps;
    private boolean stable;
    private boolean complete;
    private List<FailedStep> failedSteps;

    public PipelineStatus() {
        this.progressingSteps = new ArrayList<>();
        this.stable = true;
        this.complete = true;
        this.failedSteps = new ArrayList<>();
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
            if (deploymentLog.getStatus().equals(DeploymentLog.DeploymentStatus.FAILURE.name())) {
                this.stable = false;
            }
        } else {
            //Was instance of a RemediationLog
            if (!((RemediationLog) log).isStateRemediated()) {
                this.stable = false;
            }
            //TODO: Needs to be added to failed step
        }
    }

    public void addLatestDeploymentLog(DeploymentLog log, String step) {
        if (log.getStatus().equals(DeploymentLog.DeploymentStatus.FAILURE.name())) {
            this.complete = false;
            addFailedDeploymentLog(log, step);
        }
    }

    public void addFailedDeploymentLog(DeploymentLog log, String step) {
        this.failedSteps.add(new FailedStep(step, !log.isDeploymentComplete(), log.getBrokenTest(), log.getBrokenTestLog()));
    }
}
