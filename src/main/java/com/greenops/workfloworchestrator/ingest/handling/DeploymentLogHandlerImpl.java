package com.greenops.workfloworchestrator.ingest.handling;

import com.greenops.workfloworchestrator.datamodel.auditlog.DeploymentLog;
import com.greenops.workfloworchestrator.datamodel.event.Event;
import com.greenops.workfloworchestrator.ingest.dbclient.DbClient;
import com.greenops.workfloworchestrator.ingest.dbclient.DbKey;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

@Slf4j
@Component
public class DeploymentLogHandlerImpl implements DeploymentLogHandler {

    private DbClient dbClient;

    @Autowired
    DeploymentLogHandlerImpl(DbClient dbClient) {
        this.dbClient = dbClient;
    }

    @Override
    public boolean updateStepDeploymentLog(Event event, String stepName, String argoApplicationName, int revisionId) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var deploymentLog = dbClient.fetchLatestLog(logKey);
        if (deploymentLog != null && revisionId > -1) {
            deploymentLog.setArgoApplicationName(argoApplicationName);
            deploymentLog.setArgoRevisionId(revisionId);
            return dbClient.updateHeadInList(logKey, deploymentLog);
        }
        return true;
    }

    @Override
    public boolean initializeNewStepLog(Event event, String stepName, String gitCommitVersion) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var newLog = new DeploymentLog(DeploymentLog.DeploymentStatus.PROGRESSING.name(), false, -1, gitCommitVersion);
        return dbClient.insertValueInList(logKey, newLog);
    }

    //Remember, this is the deployment, not the step
    @Override
    public boolean markDeploymentSuccessful(Event event, String stepName) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var deploymentLog = dbClient.fetchLatestLog(logKey);
        if (deploymentLog != null) {
            deploymentLog.setDeploymentComplete(true);
            return dbClient.updateHeadInList(logKey, deploymentLog);
        }
        return true;
    }

    @Override
    public boolean markStepSuccessful(Event event, String stepName) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var deploymentLog = dbClient.fetchLatestLog(logKey);
        if (deploymentLog != null && deploymentLog.getBrokenTest() == null) {
            deploymentLog.setStatus(DeploymentLog.DeploymentStatus.SUCCESS.name());
            return dbClient.updateHeadInList(logKey, deploymentLog);
        }
        return true;
    }

    @Override
    public boolean markStepFailedWithBrokenTest(Event event, String stepName, String testName, String testLog) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var deploymentLog = dbClient.fetchLatestLog(logKey);
        if (deploymentLog != null) {
            deploymentLog.setBrokenTest(testName);
            deploymentLog.setBrokenTestLog(testLog);
            deploymentLog.setStatus(DeploymentLog.DeploymentStatus.FAILURE.name());
            return dbClient.updateHeadInList(logKey, deploymentLog);
        }
        return true;
    }

    @Override
    public boolean areParentStepsComplete(String stepName) {
        //TODO: Check redis and implement the flow for ensuring the parent steps have all been completed
        return true;
    }

    //Returning null means a failure occurred, empty string means no match exists.
    @Override
    public String makeRollbackDeploymentLog(Event event, String stepName) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var deploymentLogList = dbClient.fetchLogList(logKey);
        if (deploymentLogList == null || deploymentLogList.size() == 0) return "";
        var currentLog = deploymentLogList.get(0);
        //This means that there was probably an error during the execution of the step, and that the log was added but the re-triggering process was not completed
        if (currentLog.getStatus().equals(DeploymentLog.DeploymentStatus.PROGRESSING.name()) && currentLog.getUniqueVersionInstance() > 0) {
            return currentLog.getGitCommitVersion();
        }
        int idx = 0;
        if (currentLog.getUniqueVersionInstance() > 0) {
            while (idx < deploymentLogList.size()) {
                if (currentLog.getUniqueVersionNumber().equals(deploymentLogList.get(idx).getUniqueVersionNumber())
                        && deploymentLogList.get(idx).getUniqueVersionInstance() == 0) {
                    idx++;
                    break;
                }
                idx++;
            }
        }
        for (; idx < deploymentLogList.size(); idx++) {
            if (deploymentLogList.get(idx).getStatus().equals(DeploymentLog.DeploymentStatus.SUCCESS.name())) {
                var gitCommitVersion = deploymentLogList.get(idx).getGitCommitVersion();

                logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), event.getStepName());
                var newLog =  new DeploymentLog(
                        deploymentLogList.get(idx).getUniqueVersionNumber(),
                        deploymentLogList.get(idx).getUniqueVersionInstance() + 1,
                        DeploymentLog.DeploymentStatus.PROGRESSING.name(),
                        false,
                        deploymentLogList.get(idx).getArgoApplicationName(),
                        deploymentLogList.get(idx).getArgoRevisionId(),
                        gitCommitVersion,
                        null,
                        null
                );
                if (!dbClient.insertValueInList(logKey, newLog)) {
                    return null;
                }
                return gitCommitVersion;
            }
        }
        return "";
    }
}
