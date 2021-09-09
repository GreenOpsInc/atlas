package com.greenops.workflowtrigger.api;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.util.datamodel.auditlog.DeploymentLog;
import com.greenops.util.datamodel.auditlog.Log;
import com.greenops.util.datamodel.pipelinestatus.PipelineStatus;
import com.greenops.util.dbclient.DbClient;
import com.greenops.util.error.AtlasNonRetryableError;
import com.greenops.workflowtrigger.dbclient.DbKey;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.ArrayList;
import java.util.List;

import static com.greenops.util.dbclient.DbClient.LOG_INCREMENT;

@Slf4j
@RestController
@RequestMapping("/status")
public class PipelineStatusApi {

    private final DbClient dbClient;
    private final ObjectMapper objectMapper;

    @Autowired
    public PipelineStatusApi(DbClient dbClient, ObjectMapper objectMapper) {
        this.dbClient = dbClient;
        this.objectMapper = objectMapper;
    }

    @GetMapping(value = "{orgName}/{teamName}/pipeline/{pipelineName}/step/{stepName}/{count}")
    public ResponseEntity<List<Log>> getStepLogs(@PathVariable("orgName") String orgName,
                                                           @PathVariable("teamName") String teamName,
                                                           @PathVariable("pipelineName") String pipelineName,
                                                           @PathVariable("stepName") String stepName,
                                                           @PathVariable("count") int count) {
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
        var status = new PipelineStatus();
        var steps = dbClient.fetchStringList(DbKey.makeDbListOfStepsKey(orgName, teamName, pipelineName));
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

    private String schemaToResponsePayload(Object schema) {
        try {
            return objectMapper.writeValueAsString(schema);
        } catch (JsonProcessingException e) {
            throw new RuntimeException("Could not convert schema into response payload.", e);
        }
    }
}
