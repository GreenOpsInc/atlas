package com.greenops.workflowtrigger.api;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.util.datamodel.auditlog.DeploymentLog;
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
    public ResponseEntity<List<DeploymentLog>> getStepLogs(@PathVariable("orgName") String orgName,
                                                           @PathVariable("teamName") String teamName,
                                                           @PathVariable("pipelineName") String pipelineName,
                                                           @PathVariable("stepName") String stepName,
                                                           @PathVariable("count") int count) {
        var key = DbKey.makeDbStepKey(orgName, teamName, pipelineName, stepName);
        var increments = (int) Math.ceil(LOG_INCREMENT / (double) count);
        var deploymentLogList = new ArrayList<DeploymentLog>();
        for (int idx = 0; idx < increments; idx++) {
            var fetchedLogList = dbClient.fetchLogList(key, idx);
            if (idx == increments - 1) {
                var difference = count - ((increments - 1) * LOG_INCREMENT);
                deploymentLogList.addAll(fetchedLogList.subList(0, Math.min(difference, fetchedLogList.size())));
            } else {
                deploymentLogList.addAll(fetchedLogList);
            }
        }

        return ResponseEntity.ok(deploymentLogList);
    }

    @GetMapping(value = "{orgName}/{teamName}/pipeline/{pipelineName}")
    public ResponseEntity<String> getPipelineStatus(@PathVariable("orgName") String orgName,
                                                    @PathVariable("teamName") String teamName,
                                                    @PathVariable("pipelineName") String pipelineName) {
        var status = new PipelineStatus();
        var steps = dbClient.fetchStringList(DbKey.makeDbListOfStepsKey(orgName, teamName, pipelineName));
        for (var step : steps) {
            try {
                var logKey = DbKey.makeDbStepKey(orgName, teamName, pipelineName, step);
                var deploymentLog = dbClient.fetchLatestLog(logKey);
                status.addDeploymentLog(deploymentLog, step);

                // was rollback deployment, need to get failure logs of initial deployment
                if (deploymentLog.getUniqueVersionInstance() != 0) {
                    var logIncrement = 0;
                    var deploymentLogList = dbClient.fetchLogList(logKey, logIncrement);
                    int idx = 0;
                    while (idx < deploymentLogList.size()) {
                        if (deploymentLogList.get(idx).getUniqueVersionInstance() == 0) {
                            status.addFailedDeploymentLog(deploymentLogList.get(idx), step);
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
