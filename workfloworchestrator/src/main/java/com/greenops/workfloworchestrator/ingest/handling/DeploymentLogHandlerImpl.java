package com.greenops.workfloworchestrator.ingest.handling;

import com.greenops.util.datamodel.auditlog.DeploymentLog;
import com.greenops.util.datamodel.auditlog.RemediationLog;
import com.greenops.util.datamodel.event.Event;
import com.greenops.util.dbclient.DbClient;
import com.greenops.util.datamodel.clientmessages.ResourceGvk;
import com.greenops.workfloworchestrator.error.AtlasNonRetryableError;
import com.greenops.workfloworchestrator.ingest.dbclient.DbKey;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.stream.Collectors;

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
        var deploymentLog = dbClient.fetchLatestDeploymentLog(logKey);
        deploymentLog.setArgoApplicationName(argoApplicationName);
        deploymentLog.setArgoRevisionHash(revisionHash);
        dbClient.updateHeadInList(logKey, deploymentLog);
    }

    @Override
    public void initializeNewStepLog(Event event, String stepName, String pipelineUvn, String gitCommitVersion) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var newLog = new DeploymentLog(pipelineUvn, DeploymentLog.DeploymentStatus.PROGRESSING.name(), false, null, gitCommitVersion);
        dbClient.insertValueInList(logKey, newLog);
    }

    @Override
    public void initializeNewRemediationLog(Event event, String stepName, String pipelineUvn, List<ResourceGvk> resourceGvkList) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var latestRemediationLog = dbClient.fetchLatestRemediationLog(logKey);
        if (latestRemediationLog == null || !latestRemediationLog.getPipelineUniqueVersionNumber().equals(pipelineUvn)) {
            latestRemediationLog = new RemediationLog(pipelineUvn, 0, List.of());
        }
        var newLog = new RemediationLog(
                pipelineUvn,
                latestRemediationLog.getUniqueVersionInstance() + 1,
                resourceGvkList.stream().map(ResourceGvk::getResourceName).collect(Collectors.toList())
        );
        dbClient.insertValueInList(logKey, newLog);
    }

    //Remember, this is the deployment, not the step
    @Override
    public void markDeploymentSuccessful(Event event, String stepName) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        //TODO: Remove this line when the TriggerStep event is added
        if (stepName.equals(ROOT_STEP_NAME)) return;
        var deploymentLog = dbClient.fetchLatestDeploymentLog(logKey);
        deploymentLog.setDeploymentComplete(true);
        dbClient.updateHeadInList(logKey, deploymentLog);
    }

    @Override
    public void markStepSuccessful(Event event, String stepName) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        //TODO: Remove this line when the TriggerStep event is added
        if (stepName.equals(ROOT_STEP_NAME)) return;
        var deploymentLog = dbClient.fetchLatestDeploymentLog(logKey);
        //This check is largely redundant. Should never be the case
        if (deploymentLog.getBrokenTest() != null) {
            throw new AtlasNonRetryableError("This step has test failures. Should not be marked successful");
        }
        deploymentLog.setStatus(DeploymentLog.DeploymentStatus.SUCCESS.name());
        dbClient.updateHeadInList(logKey, deploymentLog);
    }

    @Override
    public void markStateRemediated(Event event, String stepName) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        //TODO: Remove this line when the TriggerStep event is added
        if (stepName.equals(ROOT_STEP_NAME)) return;
        var remediationLog = dbClient.fetchLatestRemediationLog(logKey);
        if (remediationLog == null) {
            log.info("No remediation log present");
            return;
        }
        remediationLog.setStateRemediated(RemediationLog.RemediationStatus.SUCCESS.name());
        dbClient.updateHeadInList(logKey, remediationLog);
    }

    @Override
    public void markStateRemediationFailed(Event event, String stepName) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        //TODO: Remove this line when the TriggerStep event is added
        if (stepName.equals(ROOT_STEP_NAME)) return;
        var remediationLog = dbClient.fetchLatestRemediationLog(logKey);
        if (remediationLog == null) {
            log.info("No remediation log present");
            return;
        }
        remediationLog.setStateRemediated(RemediationLog.RemediationStatus.FAILURE.name());
        dbClient.updateHeadInList(logKey, remediationLog);
    }

    @Override
    public void markStepFailedWithFailedDeployment(Event event, String stepName) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var deploymentLog = dbClient.fetchLatestDeploymentLog(logKey);
        deploymentLog.setDeploymentComplete(false);
        deploymentLog.setStatus(DeploymentLog.DeploymentStatus.FAILURE.name());
        dbClient.updateHeadInList(logKey, deploymentLog);
    }

    @Override
    public void markStepFailedWithBrokenTest(Event event, String stepName, String testName, String testLog) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var deploymentLog = dbClient.fetchLatestDeploymentLog(logKey);
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
            var deploymentLog = dbClient.fetchLatestDeploymentLog(logKey);
            if (deploymentLog.getUniqueVersionInstance() != 0 || !deploymentLog.getStatus().equals(DeploymentLog.DeploymentStatus.SUCCESS.name())) {
                return false;
            }
        }
        return true;
    }

    @Override
    public String getStepStatus(Event event) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), event.getStepName());
        var deploymentLog = dbClient.fetchLatestDeploymentLog(logKey);
        return deploymentLog.getStatus();
    }

    //Returning null means a failure occurred, empty string means no match exists.
    @Override
    public String makeRollbackDeploymentLog(Event event, String stepName) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var logIncrement = 0;
        var logList = dbClient.fetchLogList(logKey, logIncrement);
        if (logList == null || logList.size() == 0) return "";
        var currentLog = dbClient.fetchLatestDeploymentLog(logKey);
        //This means that there was probably an error during the execution of the step, and that the log was added but the re-triggering process was not completed
        if (currentLog.getStatus().equals(DeploymentLog.DeploymentStatus.PROGRESSING.name()) && currentLog.getUniqueVersionInstance() > 0) {
            return currentLog.getGitCommitVersion();
        }
        int idx = 0;
        //If a specific rollback UVN has already been tried but has failed, we want to skip to the first instance of that UVN and search before then.
        if (currentLog.getUniqueVersionInstance() > 0) {
            while (idx < logList.size()) {
                if (currentLog.getRollbackUniqueVersionNumber().equals(logList.get(idx).getPipelineUniqueVersionNumber())
                        && logList.get(idx).getUniqueVersionInstance() == 0) {
                    idx++;
                    break;
                }
                idx++;
                if (idx == logList.size()) {
                    logIncrement++;
                    logList = dbClient.fetchLogList(logKey, logIncrement);
                    idx = 0;
                }
            }
        } else if (currentLog.getStatus().equals(DeploymentLog.DeploymentStatus.SUCCESS.name())) {
            while (idx < logList.size()) {
                if (!(logList.get(idx) instanceof DeploymentLog)) {
                    idx++;
                    continue;
                }
                if (!currentLog.getPipelineUniqueVersionNumber().equals((logList.get(idx)).getPipelineUniqueVersionNumber())) {
                    break;
                }
                idx++;
                if (idx == logList.size()) {
                    logIncrement++;
                    logList = dbClient.fetchLogList(logKey, logIncrement);
                    idx = 0;
                }
            }
        }
        while (idx < logList.size()) {
            if (!(logList.get(idx) instanceof DeploymentLog)) {
                idx++;
                continue;
            }
            var deploymentLog = (DeploymentLog) logList.get(idx);
            if (deploymentLog.getStatus().equals(DeploymentLog.DeploymentStatus.SUCCESS.name())) {
                var gitCommitVersion = deploymentLog.getGitCommitVersion();

                logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), event.getStepName());
                var newLog = new DeploymentLog(
                        currentLog.getPipelineUniqueVersionNumber(),
                        logList.get(idx).getPipelineUniqueVersionNumber(),
                        logList.get(idx).getUniqueVersionInstance() + 1,
                        DeploymentLog.DeploymentStatus.PROGRESSING.name(),
                        false,
                        deploymentLog.getArgoApplicationName(),
                        deploymentLog.getArgoRevisionHash(),
                        gitCommitVersion,
                        null,
                        null
                );
                dbClient.insertValueInList(logKey, newLog);
                return gitCommitVersion;
            }
            idx++;
            if (idx == logList.size()) {
                logIncrement++;
                logList = dbClient.fetchLogList(logKey, logIncrement);
                idx = 0;
            }
        }
        return "";
    }

    @Override
    public String getCurrentGitCommitHash(Event event, String stepName) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var currentDeploymentLog = dbClient.fetchLatestDeploymentLog(logKey);
        if (currentDeploymentLog == null)
            throw new AtlasNonRetryableError("No deployment log found for this key, no commit hash will be found.");
        return currentDeploymentLog.getGitCommitVersion();
    }

    @Override
    public String getCurrentArgoRevisionHash(Event event, String stepName) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var currentDeploymentLog = dbClient.fetchLatestDeploymentLog(logKey);
        if (currentDeploymentLog == null)
            throw new AtlasNonRetryableError("No deployment log found for this key, no commit hash will be found.");
        return currentDeploymentLog.getArgoRevisionHash();
    }

    @Override
    public String getCurrentPipelineUvn(Event event, String stepName) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var currentDeploymentLog = dbClient.fetchLatestDeploymentLog(logKey);
        if (currentDeploymentLog == null)
            throw new AtlasNonRetryableError("No deployment log found for this key, no commit hash will be found.");
        return currentDeploymentLog.getPipelineUniqueVersionNumber();
    }

    @Override
    public String getLastSuccessfulStepGitCommitHash(Event event, String stepName) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var logIncrement = 0;
        var logList = dbClient.fetchLogList(logKey, logIncrement);
        if (logList == null || logList.size() == 0) return null;
        int idx = 0;
        while (idx < logList.size()) {
            if (!(logList.get(idx) instanceof DeploymentLog)) {
                idx++;
                continue;
            }
            var deploymentLog = (DeploymentLog) logList.get(idx);
            if (deploymentLog.getStatus().equals(DeploymentLog.DeploymentStatus.SUCCESS.name())) {
                return deploymentLog.getGitCommitVersion();
            }
            idx++;
            if (idx == logList.size()) {
                logIncrement++;
                logList = dbClient.fetchLogList(logKey, logIncrement);
                idx = 0;
            }
        }
        return null;
    }

    @Override
    public String getLastSuccessfulDeploymentGitCommitHash(Event event, String stepName) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var logIncrement = 0;
        var logList = dbClient.fetchLogList(logKey, logIncrement);
        if (logList == null || logList.size() == 0) return null;
        int idx = 0;
        while (idx < logList.size()) {
            if (!(logList.get(idx) instanceof DeploymentLog)) {
                idx++;
                continue;
            }
            var deploymentLog = (DeploymentLog) logList.get(idx);
            if (deploymentLog.isDeploymentComplete()) {
                return deploymentLog.getGitCommitVersion();
            }
            idx++;
            if (idx == logList.size()) {
                logIncrement++;
                logList = dbClient.fetchLogList(logKey, logIncrement);
                idx = 0;
            }
        }
        return null;
    }

    @Override
    public DeploymentLog getLatestDeploymentLog(Event event, String stepName) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        return dbClient.fetchLatestDeploymentLog(logKey);
    }
}
