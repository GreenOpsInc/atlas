package com.greenops.workfloworchestrator.ingest.handling;

import com.greenops.util.datamodel.auditlog.DeploymentLog;
import com.greenops.util.datamodel.event.Event;
import com.greenops.util.dbclient.DbClient;
import com.greenops.workfloworchestrator.error.AtlasNonRetryableError;
import com.greenops.workfloworchestrator.ingest.dbclient.DbKey;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.util.List;

import static com.greenops.workfloworchestrator.datamodel.pipelinedata.StepData.ROOT_STEP_NAME;

@Slf4j
@Component
public class DeploymentLogHandlerImpl implements DeploymentLogHandler {

    private DbClient dbClient;

    @Autowired
    DeploymentLogHandlerImpl(DbClient dbClient) {
        this.dbClient = dbClient;
    }

    @Override
    public void updateStepDeploymentLog(Event event, String stepName, String argoApplicationName, String revisionHash) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var deploymentLog = dbClient.fetchLatestLog(logKey);
        deploymentLog.setArgoApplicationName(argoApplicationName);
        deploymentLog.setArgoRevisionHash(revisionHash);
        dbClient.updateHeadInList(logKey, deploymentLog);
    }

    @Override
    public void initializeNewStepLog(Event event, String stepName, String pipelineUvn, String argoRevisionHash, String gitCommitVersion) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var newLog = new DeploymentLog(pipelineUvn, DeploymentLog.DeploymentStatus.PROGRESSING.name(), false, argoRevisionHash, gitCommitVersion);
        dbClient.insertValueInList(logKey, newLog);
    }

    //Remember, this is the deployment, not the step
    @Override
    public void markDeploymentSuccessful(Event event, String stepName) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        //TODO: Remove this line when the TriggerStep event is added
        if (stepName.equals(ROOT_STEP_NAME)) return;
        var deploymentLog = dbClient.fetchLatestLog(logKey);
        deploymentLog.setDeploymentComplete(true);
        dbClient.updateHeadInList(logKey, deploymentLog);
    }

    @Override
    public void markStepSuccessful(Event event, String stepName) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        //TODO: Remove this line when the TriggerStep event is added
        if (stepName.equals(ROOT_STEP_NAME)) return;
        var deploymentLog = dbClient.fetchLatestLog(logKey);
        //This check is largely redundant. Should never be the case
        if (deploymentLog.getBrokenTest() != null) {
            throw new AtlasNonRetryableError("This step has test failures. Should not be marked successful");
        }
        deploymentLog.setStatus(DeploymentLog.DeploymentStatus.SUCCESS.name());
        dbClient.updateHeadInList(logKey, deploymentLog);

    }

    @Override
    public void markStepFailedWithFailedDeployment(Event event, String stepName) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var deploymentLog = dbClient.fetchLatestLog(logKey);
        deploymentLog.setDeploymentComplete(false);
        deploymentLog.setStatus(DeploymentLog.DeploymentStatus.FAILURE.name());
        dbClient.updateHeadInList(logKey, deploymentLog);
    }

    @Override
    public void markStepFailedWithBrokenTest(Event event, String stepName, String testName, String testLog) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var deploymentLog = dbClient.fetchLatestLog(logKey);
        deploymentLog.setBrokenTest(testName);
        deploymentLog.setBrokenTestLog(testLog);
        deploymentLog.setStatus(DeploymentLog.DeploymentStatus.FAILURE.name());
        dbClient.updateHeadInList(logKey, deploymentLog);
    }

    @Override
    public boolean areParentStepsComplete(Event event, List<String> parentSteps) {
        for (var parentStepName : parentSteps) {
            if (parentStepName.equals(ROOT_STEP_NAME)) continue;
            var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), parentStepName);
            var deploymentLog = dbClient.fetchLatestLog(logKey);
            if (deploymentLog.getUniqueVersionInstance() != 0 || !deploymentLog.getStatus().equals(DeploymentLog.DeploymentStatus.SUCCESS.name())) {
                return false;
            }
        }
        return true;
    }

    //Returning null means a failure occurred, empty string means no match exists.
    @Override
    public String makeRollbackDeploymentLog(Event event, String stepName) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var logIncrement = 0;
        var deploymentLogList = dbClient.fetchLogList(logKey, logIncrement);
        if (deploymentLogList == null || deploymentLogList.size() == 0) return "";
        var currentLog = deploymentLogList.get(0);
        //This means that there was probably an error during the execution of the step, and that the log was added but the re-triggering process was not completed
        if (currentLog.getStatus().equals(DeploymentLog.DeploymentStatus.PROGRESSING.name()) && currentLog.getUniqueVersionInstance() > 0) {
            return currentLog.getGitCommitVersion();
        }
        int idx = 0;
        if (currentLog.getUniqueVersionInstance() > 0) {
            while (idx < deploymentLogList.size()) {
                if (currentLog.getRollbackUniqueVersionNumber().equals(deploymentLogList.get(idx).getPipelineUniqueVersionNumber())
                        && deploymentLogList.get(idx).getUniqueVersionInstance() == 0) {
                    idx++;
                    break;
                }
                idx++;
                if (idx == deploymentLogList.size()) {
                    logIncrement++;
                    deploymentLogList = dbClient.fetchLogList(logKey, logIncrement);
                    idx = 0;
                }
            }
        }
        while (idx < deploymentLogList.size()) {
            if (deploymentLogList.get(idx).getStatus().equals(DeploymentLog.DeploymentStatus.SUCCESS.name())) {
                var gitCommitVersion = deploymentLogList.get(idx).getGitCommitVersion();

                logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), event.getStepName());
                var newLog = new DeploymentLog(
                        currentLog.getPipelineUniqueVersionNumber(),
                        deploymentLogList.get(idx).getPipelineUniqueVersionNumber(),
                        deploymentLogList.get(idx).getUniqueVersionInstance() + 1,
                        DeploymentLog.DeploymentStatus.PROGRESSING.name(),
                        false,
                        deploymentLogList.get(idx).getArgoApplicationName(),
                        deploymentLogList.get(idx).getArgoRevisionHash(),
                        gitCommitVersion,
                        null,
                        null
                );
                dbClient.insertValueInList(logKey, newLog);
                return gitCommitVersion;
            }
            idx++;
            if (idx == deploymentLogList.size()) {
                logIncrement++;
                deploymentLogList = dbClient.fetchLogList(logKey, logIncrement);
                idx = 0;
            }
        }
        return "";
    }

    @Override
    public String getCurrentGitCommitHash(Event event, String stepName) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var currentDeploymentLog = dbClient.fetchLatestLog(logKey);
        if (currentDeploymentLog == null)
            throw new AtlasNonRetryableError("No deployment log found for this key, no commit hash will be found.");
        return currentDeploymentLog.getGitCommitVersion();
    }

    @Override
    public String getCurrentArgoRevisionHash(Event event, String stepName) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var currentDeploymentLog = dbClient.fetchLatestLog(logKey);
        if (currentDeploymentLog == null)
            throw new AtlasNonRetryableError("No deployment log found for this key, no commit hash will be found.");
        return currentDeploymentLog.getArgoRevisionHash();
    }

    @Override
    public String getCurrentPipelineUvn(Event event, String stepName) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var currentDeploymentLog = dbClient.fetchLatestLog(logKey);
        if (currentDeploymentLog == null)
            throw new AtlasNonRetryableError("No deployment log found for this key, no commit hash will be found.");
        return currentDeploymentLog.getPipelineUniqueVersionNumber();
    }

    @Override
    public String getLastSuccessfulStepGitCommitHash(Event event, String stepName) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var logIncrement = 0;
        var deploymentLogList = dbClient.fetchLogList(logKey, logIncrement);
        if (deploymentLogList == null || deploymentLogList.size() == 0) return null;
        int idx = 0;
        while (idx < deploymentLogList.size()) {
            if (deploymentLogList.get(idx).getStatus().equals(DeploymentLog.DeploymentStatus.SUCCESS.name())) {
                return deploymentLogList.get(idx).getGitCommitVersion();
            }
            idx++;
            if (idx == deploymentLogList.size()) {
                logIncrement++;
                deploymentLogList = dbClient.fetchLogList(logKey, logIncrement);
                idx = 0;
            }
        }
        return null;
    }

    @Override
    public String getLastSuccessfulDeploymentGitCommitHash(Event event, String stepName) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var logIncrement = 0;
        var deploymentLogList = dbClient.fetchLogList(logKey, logIncrement);
        if (deploymentLogList == null || deploymentLogList.size() == 0) return null;
        int idx = 0;
        while (idx < deploymentLogList.size()) {
            if (deploymentLogList.get(idx).isDeploymentComplete()) {
                return deploymentLogList.get(idx).getGitCommitVersion();
            }
            idx++;
            if (idx == deploymentLogList.size()) {
                logIncrement++;
                deploymentLogList = dbClient.fetchLogList(logKey, logIncrement);
                idx = 0;
            }
        }
        return null;
    }
}
