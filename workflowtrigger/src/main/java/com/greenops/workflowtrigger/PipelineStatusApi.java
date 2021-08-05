package com.greenops.workflowtrigger;

import com.greenops.util.datamodel.auditlog.DeploymentLog;
import com.greenops.util.dbclient.DbClient;
import com.greenops.workflowtrigger.dbclient.DbKey;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
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

    @Autowired
    public PipelineStatusApi(DbClient dbClient) {
        this.dbClient = dbClient;
    }

    @GetMapping(value = "{orgName}/{teamName}/pipeline/{pipelineName}/step/{stepName}/{count}")
    public ResponseEntity<List<DeploymentLog>> getStepLogs(@PathVariable("orgName") String orgName,
                                                              @PathVariable("teamName") String teamName,
                                                              @PathVariable("pipelineName") String pipelineName,
                                                              @PathVariable("stepName") String stepName,
                                                              @PathVariable("count") int count) {
        var key = DbKey.makeDbStepKey(orgName, teamName, pipelineName, stepName);
        var increments = (int)Math.ceil(LOG_INCREMENT / (double)count);
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
}
