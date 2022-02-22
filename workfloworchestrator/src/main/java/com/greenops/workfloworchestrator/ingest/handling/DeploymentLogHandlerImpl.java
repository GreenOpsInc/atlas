package com.greenops.workfloworchestrator.ingest.handling;

import com.greenops.util.datamodel.auditlog.DeploymentLog;
import com.greenops.util.datamodel.auditlog.Log;
import com.greenops.util.datamodel.auditlog.PipelineInfo;
import com.greenops.util.datamodel.auditlog.RemediationLog;
import com.greenops.util.datamodel.event.Event;
import com.greenops.util.datamodel.event.FailureEvent;
import com.greenops.util.dbclient.DbClient;
import com.greenops.util.datamodel.clientmessages.ResourceGvk;
import com.greenops.util.error.AtlasNonRetryableError;
import com.greenops.workfloworchestrator.ingest.dbclient.DbKey;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.stream.Collectors;

import static com.greenops.util.datamodel.event.PipelineTriggerEvent.PIPELINE_TRIGGER_EVENT_CLASS_NAME;
import static com.greenops.util.datamodel.event.PipelineTriggerEvent.ROOT_STEP_NAME;

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
        var newLog = new DeploymentLog(pipelineUvn, Log.LogStatus.PROGRESSING.name(), false, null, gitCommitVersion);
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
        var deploymentLog = dbClient.fetchLatestDeploymentLog(logKey);
        deploymentLog.setDeploymentComplete(true);
        dbClient.updateHeadInList(logKey, deploymentLog);
    }

    @Override
    public void markStepSuccessful(Event event, String stepName) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var deploymentLog = dbClient.fetchLatestDeploymentLog(logKey);
        //This check is largely redundant. Should never be the case
        if (deploymentLog.getBrokenTest() != null) {
            throw new AtlasNonRetryableError("This step has test failures. Should not be marked successful");
        }
        deploymentLog.setStatus(Log.LogStatus.SUCCESS.name());
        dbClient.updateHeadInList(logKey, deploymentLog);
    }

    @Override
    public void markStateRemediated(Event event, String stepName) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var remediationLog = dbClient.fetchLatestRemediationLog(logKey);
        if (remediationLog == null) {
            log.info("No remediation log present");
            return;
        }
        remediationLog.setStatus(Log.LogStatus.SUCCESS.name());
        dbClient.updateHeadInList(logKey, remediationLog);
    }

    @Override
    public void markStateRemediationFailed(Event event, String stepName) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var remediationLog = dbClient.fetchLatestRemediationLog(logKey);
        if (remediationLog == null) {
            log.info("No remediation log present");
            return;
        }
        remediationLog.setStatus(Log.LogStatus.FAILURE.name());
        dbClient.updateHeadInList(logKey, remediationLog);
    }

    @Override
    public void markStepFailedWithFailedDeployment(Event event, String stepName) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var deploymentLog = dbClient.fetchLatestDeploymentLog(logKey);
        deploymentLog.setDeploymentComplete(false);
        deploymentLog.setStatus(Log.LogStatus.FAILURE.name());
        dbClient.updateHeadInList(logKey, deploymentLog);
    }

    @Override
    public void markStepFailedWithBrokenTest(Event event, String stepName, String testName, String testLog) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var deploymentLog = dbClient.fetchLatestDeploymentLog(logKey);
        deploymentLog.setBrokenTest(testName);
        deploymentLog.setBrokenTestLog(testLog);
        deploymentLog.setStatus(Log.LogStatus.FAILURE.name());
        dbClient.updateHeadInList(logKey, deploymentLog);
    }

    @Override
    public void markStepFailedWithProcessingError(FailureEvent event, String stepName, String error) {
        log.info("Marking step failed with processing error");
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var log = dbClient.fetchLatestLog(logKey);
        if (log == null) {
            markPipelineFailedWithProcessingError(event, error);
            return;
        } else if (log instanceof DeploymentLog) {
            ((DeploymentLog) log).setBrokenTest("Processing Error");
            ((DeploymentLog) log).setBrokenTestLog(error);
        }
        log.setStatus(Log.LogStatus.FAILURE.name());
        dbClient.updateHeadInList(logKey, log);
    }

    //When the deployment log for a step doesn't exist, just mark the entire pipeline ambiguously and allow the user to determine the origin.
    private void markPipelineFailedWithProcessingError(FailureEvent event, String error) {
        log.info("No step found for the error, adding in to the pipeline metadata");
        var key = DbKey.makeDbPipelineInfoKey(event.getOrgName(), event.getTeamName(), event.getPipelineName());
        var pipelineInfo = dbClient.fetchLatestPipelineInfo(key);
        if (event.getStatusCode().equals(PIPELINE_TRIGGER_EVENT_CLASS_NAME) || pipelineInfo == null) {
            pipelineInfo = new PipelineInfo(event.getPipelineUvn(), List.of(error), List.of());
            dbClient.insertValueInList(key, pipelineInfo);
        } else {
            pipelineInfo.addError(error);
            dbClient.updateHeadInList(key, pipelineInfo);
        }
    }

    @Override
    public boolean areParentStepsComplete(Event event, List<String> parentSteps) {
        for (var parentStepName : parentSteps) {
            if (parentStepName.equals(ROOT_STEP_NAME)) continue;
            var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), parentStepName);
            var deploymentLog = dbClient.fetchLatestDeploymentLog(logKey);
            if (deploymentLog.getUniqueVersionInstance() != 0 || !deploymentLog.getStatus().equals(Log.LogStatus.SUCCESS.name())) {
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
    public String makeRollbackDeploymentLog(Event event, String stepName, int rollbackLimit, boolean dryRun) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var logIncrement = 0;
        var logList = dbClient.fetchLogList(logKey, logIncrement);
        if (logList == null || logList.size() == 0) return "";
        var currentLog = dbClient.fetchLatestDeploymentLog(logKey);
        if (currentLog == null) {
            throw new AtlasNonRetryableError("A rollback was triggered, but no previous logs could be found for the step.");
        }
        if (currentLog.getUniqueVersionInstance() >= rollbackLimit) {
            log.info("Met the limit for rollbacks for this step, so returning an invalid commit hash to stop further processing.");
            return "";
        }
        //This means that there was probably an error during the execution of the step, and that the log was added but the re-triggering process was not completed
        if (currentLog.getStatus().equals(Log.LogStatus.PROGRESSING.name()) && currentLog.getUniqueVersionInstance() > 0) {
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
        } else if (currentLog.getStatus().equals(Log.LogStatus.SUCCESS.name())) {
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
            if (deploymentLog.getStatus().equals(Log.LogStatus.SUCCESS.name()) && deploymentLog.getUniqueVersionInstance() == 0) {
                var gitCommitVersion = deploymentLog.getGitCommitVersion();

                logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), event.getStepName());
                var newLog = new DeploymentLog(
                        currentLog.getPipelineUniqueVersionNumber(),
                        logList.get(idx).getPipelineUniqueVersionNumber(),
                        currentLog.getUniqueVersionInstance() + 1,
                        Log.LogStatus.PROGRESSING.name(),
                        false,
                        deploymentLog.getArgoApplicationName(),
                        deploymentLog.getArgoRevisionHash(),
                        gitCommitVersion,
                        null,
                        null
                );
                if (!dryRun) {
                    dbClient.insertValueInList(logKey, newLog);
                }
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
        if (currentDeploymentLog == null) return null;
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
            if (deploymentLog.getStatus().equals(Log.LogStatus.SUCCESS.name())) {
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
