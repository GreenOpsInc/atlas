package com.greenops.verificationtool.ingest.handling;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.util.datamodel.auditlog.DeploymentLog;
import com.greenops.util.datamodel.auditlog.Log;
import com.greenops.util.datamodel.event.Event;
import com.greenops.util.datamodel.event.PipelineCompletionEvent;
import com.greenops.util.datamodel.event.TriggerStepEvent;
import com.greenops.verificationtool.datamodel.verification.DAG;
import com.greenops.verificationtool.datamodel.verification.Vertex;
import com.greenops.verificationtool.ingest.apiclient.workflowtrigger.WorkflowTriggerApi;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Component;

import java.util.List;

@Component
public class StepVerificationHandlerImpl implements StepVerificationHandler {
    private final String POPULATED = "POPULATED";
    private final String PROGRESSING = "PROGRESSING";
    private final String SUCCESS = "SUCCESS";
    private final String ClientCompletionEvent = "ClientCompletionEvent";
    private final String TestCompletionEvent = "TestCompletionEvent";
    private final WorkflowTriggerApi workflowTriggerApi;
    private final ObjectMapper objectMapper;

    @Autowired
    public StepVerificationHandlerImpl(WorkflowTriggerApi workflowTriggerApi, @Qualifier("objectMapper") ObjectMapper objectMapper) {
        this.workflowTriggerApi = workflowTriggerApi;
        this.objectMapper = objectMapper;
    }

    private List<Log> getLogs(String orgName, String pipelineName, String teamName, String stepName) {
        var stepLevelStatusStr = this.workflowTriggerApi.getStepLevelStatus(orgName, pipelineName, teamName, stepName, 15);
        System.out.println("Step Level Status: " + stepLevelStatusStr);
        try {
            return this.objectMapper.readValue(stepLevelStatusStr, new TypeReference<List<Log>>() {});
        } catch (JsonProcessingException e) {
            throw new RuntimeException("Unable to serialize step level logs into List<Log> Object " + e);
        }
    }

    private Boolean verifyHelper(Event event, DAG dag, List<Log> stepLevelLogs){
        var prevVertices = dag.getPreviousVertices(event);
        if (stepLevelLogs == null || !stepLevelLogs.get(0).getPipelineUniqueVersionNumber().equals(event.getPipelineUvn())) {
            for (Vertex prevVertex : prevVertices) {
                if (prevVertex.getStepName().equals(event.getStepName())) {
                    return false;
                }
            }
            return true;
        }
        var log = stepLevelLogs.get(0);
        for (Vertex prevVertex : prevVertices) {
            // Last Event in a Step
            if ((event instanceof TriggerStepEvent || event instanceof PipelineCompletionEvent) && (prevVertex.getEventType().equals(this.ClientCompletionEvent) || (prevVertex.getEventType().equals(this.TestCompletionEvent)))) {
                if (log.getStatus().equals(this.SUCCESS) && event.getPipelineUvn().equals(log.getPipelineUniqueVersionNumber())) {
                    return true;
                }
            }
            if (prevVertex.getEventType().equals(this.ClientCompletionEvent) && !((DeploymentLog) log).isDeploymentComplete()) {
                return false;
            }
            if (prevVertex.getEventType().equals(this.TestCompletionEvent) && dag.isTestCompletionBefore(prevVertex)) {
                if (!((DeploymentLog) log).getArgoApplicationName().equals("") || !((DeploymentLog) log).getArgoRevisionHash().equals("")) {
                    return false;
                }
            }
        }
        return event.getPipelineUvn().equals(log.getPipelineUniqueVersionNumber())
                && log.getStatus().equals(this.PROGRESSING);
    }

    @Override
    public Boolean verify(Event event, DAG dag) {
        if (event instanceof TriggerStepEvent || event instanceof PipelineCompletionEvent) {
            var vertices = dag.getPreviousVertices(event);
            for (var vertex : vertices) {
                if (vertex.getEventType().equals(this.ClientCompletionEvent) || vertex.getEventType().equals(this.TestCompletionEvent)) {
                    var logs = getLogs(event.getOrgName(), event.getPipelineName(), event.getTeamName(), vertex.getStepName());
                    if (verifyHelper(event, dag, logs)) {
                        System.out.println(event.getOrgName() + " " + event.getPipelineName() + " " + event.getTeamName() + " " + vertex.getEventType() + " Step Level Completion Verification Passed!");
                        return true;
                    } else {
                        System.out.println(event.getOrgName() + " " + event.getPipelineName() + " " + event.getTeamName() + " " + vertex.getEventType() + " Step Level Completion Verification Failed!");
                        return false;
                    }
                }
            }
        }
        var stepLevelStatus = getLogs(event.getOrgName(), event.getPipelineName(), event.getTeamName(), event.getStepName());
        if (verifyHelper(event, dag, stepLevelStatus)) {
            System.out.println(event.getOrgName() + " " + event.getPipelineName() + " " + event.getTeamName() + " " + event.getClass().getName() + " Step Level Verification Passed!");
            return true;
        } else {
            return false;
        }
    }

    @Override
    public Boolean verifyExpected(Event event, Log expectedLog) {
        var stepLevelStatus = getLogs(event.getOrgName(), event.getPipelineName(), event.getTeamName(), event.getStepName());
        if ((expectedLog == null && stepLevelStatus == null) || (expectedLog == null && stepLevelStatus.get(0) == null)) {
            return true;
        }
        if ((stepLevelStatus == null && expectedLog != null) || (stepLevelStatus != null && expectedLog == null)) {
            return false;
        }
        if ((stepLevelStatus.get(0) == null && expectedLog != null) || (stepLevelStatus.get(0) != null && expectedLog == null)) {
            return false;
        }
        var log = stepLevelStatus.get(0);
        if (!expectedLog.getPipelineUniqueVersionNumber().equals(log.getPipelineUniqueVersionNumber()) &&
                (expectedLog.getPipelineUniqueVersionNumber().equals(this.POPULATED) && log.getPipelineUniqueVersionNumber().equals("")))
            return false;
        if (!expectedLog.getStatus().equals(log.getStatus()) && (expectedLog.getStatus().equals(this.POPULATED) && log.getStatus().equals("")))
            return false;
        if (log instanceof DeploymentLog && ((DeploymentLog) expectedLog).isDeploymentComplete() != ((DeploymentLog) expectedLog).isDeploymentComplete())
            return false;
        if (log instanceof DeploymentLog && !((DeploymentLog) expectedLog).getArgoApplicationName().equals(((DeploymentLog) log).getArgoApplicationName()) &&
                (((DeploymentLog) expectedLog).getArgoApplicationName().equals(this.POPULATED) && ((DeploymentLog) log).getArgoApplicationName().equals("")))
            return false;
        return !(log instanceof DeploymentLog) || ((DeploymentLog) expectedLog).getArgoRevisionHash().equals(((DeploymentLog) log).getArgoRevisionHash()) ||
                (!((DeploymentLog) expectedLog).getArgoRevisionHash().equals(this.POPULATED) || !((DeploymentLog) log).getArgoRevisionHash().equals(""));
    }
}