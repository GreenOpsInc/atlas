package com.greenops.workflowtrigger.api;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.util.datamodel.auditlog.DeploymentLog;
import com.greenops.util.datamodel.auditlog.Log;
import com.greenops.util.datamodel.pipelinestatus.PipelineStatus;
import com.greenops.util.dbclient.DbClient;
import com.greenops.util.error.AtlasNonRetryableError;
import com.greenops.workflowtrigger.dbclient.DbKey;
import com.greenops.workflowtrigger.validator.RequestSchemaValidator;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.ArrayList;
import java.util.List;

import static com.greenops.util.dbclient.DbClient.LOG_INCREMENT;

@Slf4j
@RestController
@RequestMapping("/status")
public class PipelineStatusApi {

    private final DbClient dbClient;
    private final ObjectMapper objectMapper;
    private final RequestSchemaValidator requestSchemaValidator;

    @Autowired
    public PipelineStatusApi(DbClient dbClient, ObjectMapper objectMapper, RequestSchemaValidator requestSchemaValidator) {
        this.dbClient = dbClient;
        this.objectMapper = objectMapper;
        this.requestSchemaValidator = requestSchemaValidator;
    }

    @GetMapping(value = "{orgName}/{teamName}/pipeline/{pipelineName}/step/{stepName}/{count}")
    public ResponseEntity<List<Log>> getStepLogs(@PathVariable("orgName") String orgName,
                                                 @PathVariable("teamName") String teamName,
                                                 @PathVariable("pipelineName") String pipelineName,
                                                 @PathVariable("stepName") String stepName,
                                                 @PathVariable("count") int count) {
        if (!requestSchemaValidator.checkAuthentication()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        }
        var key = DbKey.makeDbStepKey(orgName, teamName, pipelineName, stepName);
        var increments = (int) Math.ceil(LOG_INCREMENT / (double) count);
        var logList = new ArrayList<Log>();
        for (int idx = 0; idx < increments; idx++) {
            var fetchedLogList = dbClient.fetchLogList(key, idx);
            if (idx == increments - 1) {
                var difference = count - ((increments - 1) * LOG_INCREMENT);
                logList.addAll(fetchedLogList.subList(0, Math.min(difference, fetchedLogList.size())));
            } else {
                logList.addAll(fetchedLogList);
            }
        }

        return ResponseEntity.ok(logList);
    }

    @GetMapping(value = "{orgName}/{teamName}/pipeline/{pipelineName}/{pipelineUvn}")
    public ResponseEntity<String> getPipelineStatus(@PathVariable("orgName") String orgName,
                                                    @PathVariable("teamName") String teamName,
                                                    @PathVariable("pipelineName") String pipelineName,
                                                    @PathVariable("pipelineUvn") String pipelineUvn) {
        if (!requestSchemaValidator.checkAuthentication()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        }
        var status = new PipelineStatus();
        var steps = dbClient.fetchStringList(DbKey.makeDbListOfStepsKey(orgName, teamName, pipelineName));
        if (steps == null) return ResponseEntity.ok("");
        for (var step : steps) {
            try {
                //Get pipeline UVN if not specified
                var logKey = DbKey.makeDbStepKey(orgName, teamName, pipelineName, step);
                Log log = null;
                if (pipelineUvn.equals("LATEST")) {
                    log = dbClient.fetchLatestLog(logKey);
                    if (log == null) {
                        return ResponseEntity.badRequest().body("No deployment log exists.");
                    }
                    pipelineUvn = log.getPipelineUniqueVersionNumber();
                }

                //TODO: This iteration is in enough places where it should be extracted as a dbClient method
                //Get most recent log (deployment or remediation) with desired pipeline UVN
                var logIncrement = 0;
                var logList = dbClient.fetchLogList(logKey, logIncrement);
                int idx = 0;
                while (idx < logList.size()) {
                    if (logList.get(idx).getPipelineUniqueVersionNumber().equals(pipelineUvn)) {
                        log = logList.get(idx);
                        break;
                    }
                    idx++;
                    if (idx == logList.size()) {
                        logIncrement++;
                        logList = dbClient.fetchLogList(logKey, logIncrement);
                        idx = 0;
                    }
                }

                if (log == null) {
                    status.markIncomplete();
                    continue;
                }
                if (log instanceof DeploymentLog && ((DeploymentLog) log).getStatus().equals(Log.LogStatus.PROGRESSING.name())) {
                    status.addProgressingStep(step);
                    continue;
                }
                if (log instanceof DeploymentLog && ((DeploymentLog) log).getStatus().equals(Log.LogStatus.CANCELLED.name())) {
                    status.markCancelled();
                    continue;
                }
                //Determines if the step is stable
                status.addLatestLog(log);
                //Determines if the step is complete
                if (log instanceof DeploymentLog) {
                    status.addLatestDeploymentLog((DeploymentLog) log, step);
                } else {
                    log = null;
                    while (idx < logList.size()) {
                        if (logList.get(idx).getPipelineUniqueVersionNumber().equals(pipelineUvn) && (logList.get(idx) instanceof DeploymentLog)) {
                            log = logList.get(idx);
                            status.addLatestDeploymentLog((DeploymentLog) log, step);
                            break;
                        }
                        idx++;
                        if (idx == logList.size()) {
                            logIncrement++;
                            logList = dbClient.fetchLogList(logKey, logIncrement);
                            idx = 0;
                        }
                    }
                    if (log == null) {
                        status.markIncomplete();
                        continue;
                    }
                }

                // was rollback deployment, need to get failure logs of initial deployment
                if (log.getUniqueVersionInstance() != 0) {
                    logIncrement = 0;
                    logList = dbClient.fetchLogList(logKey, logIncrement);
                    idx = 0;
                    while (idx < logList.size()) {
                        if (logList.get(idx).getUniqueVersionInstance() == 0 && (logList.get(idx) instanceof DeploymentLog)) {
                            status.addFailedDeploymentLog((DeploymentLog) logList.get(idx), step);
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
            } catch (AtlasNonRetryableError e) {
                status.markIncomplete();
            }
        }
        return ResponseEntity.ok()
                .contentType(MediaType.APPLICATION_JSON)
                .body(schemaToResponsePayload(status));
    }

    @DeleteMapping(value = "{orgName}/{teamName}/pipelineRun/{pipelineName}")
    public ResponseEntity<Void> cancelLatestPipeline(@PathVariable("orgName") String orgName,
                                                     @PathVariable("teamName") String teamName,
                                                     @PathVariable("pipelineName") String pipelineName) {
        if (!requestSchemaValidator.checkAuthentication()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        }
        var latestUvn = "";
        var steps = dbClient.fetchStringList(DbKey.makeDbListOfStepsKey(orgName, teamName, pipelineName));
        for (var stepName : steps) {
            var key = DbKey.makeDbStepKey(orgName, teamName, pipelineName, stepName);
            var latestLog = dbClient.fetchLatestLog(key);
            if (latestUvn.isEmpty()) {
                //Step list is ordered, so if the very first log is nonexistant, it can't have deployed anywhere else
                if (latestLog == null) {
                    return ResponseEntity.ok().build();
                }
                latestUvn = latestLog.getPipelineUniqueVersionNumber();
            }
            if (latestLog == null || !latestLog.getPipelineUniqueVersionNumber().equals(latestUvn)) {
                var newCancelledLog = new DeploymentLog(latestUvn, Log.LogStatus.CANCELLED.name(), false, null, null);
                dbClient.insertValueInList(key, newCancelledLog);
            } else {
                latestLog.setStatus(Log.LogStatus.CANCELLED.name());
                dbClient.updateHeadInList(key, latestLog);
            }
        }
        return ResponseEntity.ok().build();
    }

    private String schemaToResponsePayload(Object schema) {
        try {
            return objectMapper.writeValueAsString(schema);
        } catch (JsonProcessingException e) {
            throw new RuntimeException("Could not convert schema into response payload.", e);
        }
    }
}
