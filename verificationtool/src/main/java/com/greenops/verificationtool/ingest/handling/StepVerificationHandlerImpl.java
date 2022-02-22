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
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Component;

import java.util.List;

import static com.greenops.verificationtool.datamodel.status.VerificationStatusImpl.COMPLETE;
import static com.greenops.verificationtool.datamodel.status.VerificationStatusImpl.FAILURE;
import static com.greenops.verificationtool.datamodel.status.VerificationStatusImpl.EVENT_COMPLETION_FAILED;
import static com.greenops.verificationtool.datamodel.status.VerificationStatusImpl.FAILURE_EVENT_RECEIVED;

@Slf4j
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
            return this.objectMapper.readValue(stepLevelStatusStr, new TypeReference<List<Log>>() {
            });
        } catch (JsonProcessingException e) {
            throw new RuntimeException("Unable to serialize step level logs into List<Log> Object " + e);
        }
    }

    private Boolean verifyHelper(Event event, DAG dag, List<Log> stepLevelLogs) {
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

        // handling rollback_limit
        if((log.getStatus().equals(FAILURE) && (!((DeploymentLog) log).isDeploymentComplete()
                || (((DeploymentLog) log).isDeploymentComplete()) && !((DeploymentLog) log).getArgoApplicationName().equals("") && !((DeploymentLog) log).getArgoRevisionHash().equals("")))){
            return true;
        }
        for (Vertex prevVertex : prevVertices) {
            // Last Event in a Step
            if (event instanceof TriggerStepEvent && (prevVertex.getEventType().equals(this.ClientCompletionEvent) || (prevVertex.getEventType().equals(this.TestCompletionEvent)))) {
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
        if (event instanceof TriggerStepEvent) {
            var vertices = dag.getPreviousVertices(event);
            for (var vertex : vertices) {
                if (vertex.getEventType().equals(this.ClientCompletionEvent) || vertex.getEventType().equals(this.TestCompletionEvent)) {
                    var logs = getLogs(event.getOrgName(), event.getPipelineName(), event.getTeamName(), vertex.getStepName());
                    if (verifyHelper(event, dag, logs)) {
                        return true;
                    } else {
                        return false;
                    }
                }
            }
        }
        var stepLevelStatus = getLogs(event.getOrgName(), event.getPipelineName(), event.getTeamName(), event.getStepName());
        if (event instanceof PipelineCompletionEvent) {
            return verifyPipelineCompletion(event, stepLevelStatus);
        }
        return verifyHelper(event, dag, stepLevelStatus);
    }

    @Override
    public int verifyExpected(Event event, List<Log> logs) {
        var stepLevelStatus = getLogs(event.getOrgName(), event.getPipelineName(), event.getTeamName(), event.getStepName());
        if (logs == null && stepLevelStatus == null) {
            return 1;
        }
        if ((stepLevelStatus == null && logs != null) || (stepLevelStatus != null && logs == null)) {
            return 0;
        }
        if(logs.size() == stepLevelStatus.size()) {
            for(int i=0; i<logs.size(); i++) {
                var log = stepLevelStatus.get(i);
                var expectedLog = logs.get(i);
                if (!expectedLog.getPipelineUniqueVersionNumber().equals(log.getPipelineUniqueVersionNumber()) &&
                        (expectedLog.getPipelineUniqueVersionNumber().equals(this.POPULATED) && log.getPipelineUniqueVersionNumber().equals("")))
                    return 0;
                if (!expectedLog.getStatus().equals(log.getStatus()) && !(expectedLog.getStatus().equals(this.POPULATED) && !log.getStatus().equals("")))
                    return 0;
                if (log instanceof DeploymentLog && ((DeploymentLog) expectedLog).isDeploymentComplete() != ((DeploymentLog) expectedLog).isDeploymentComplete())
                    return 0;
                if (log instanceof DeploymentLog && (!((DeploymentLog) expectedLog).getArgoApplicationName().equals(((DeploymentLog) log).getArgoApplicationName()) &&
                        !(((DeploymentLog) expectedLog).getArgoApplicationName().equals(this.POPULATED) && !((DeploymentLog) log).getArgoApplicationName().equals(""))))
                    return 0;
                if (log instanceof DeploymentLog && (!((DeploymentLog) expectedLog).getArgoRevisionHash().equals(((DeploymentLog) log).getArgoRevisionHash()) &&
                        !(((DeploymentLog) expectedLog).getArgoRevisionHash().equals(this.POPULATED) && !((DeploymentLog) log).getArgoRevisionHash().equals("")))) {
                    return 0;
                }
                if (log instanceof DeploymentLog && (!((DeploymentLog) expectedLog).getBrokenTest().equals(((DeploymentLog) log).getBrokenTest()) &&
                        !(((DeploymentLog) expectedLog).getBrokenTest().equals(this.POPULATED) && !((DeploymentLog) log).getBrokenTest().equals("")))) {
                    return 0;
                }
                if (log instanceof DeploymentLog && (!((DeploymentLog) expectedLog).getBrokenTestLog().equals(((DeploymentLog) log).getBrokenTestLog()) &&
                        !(((DeploymentLog) expectedLog).getBrokenTestLog().equals(this.POPULATED) && !((DeploymentLog) log).getBrokenTestLog().equals("")))) {
                    return 0;
                }
            }
            return 1;
        }
        return 2;
    }

    private Boolean verifyPipelineCompletion(Event event, List<Log> logs) {
        if (((PipelineCompletionEvent) event).getStatus().equals(FAILURE_EVENT_RECEIVED)) {
            return true;
        } else if (((PipelineCompletionEvent) event).getStatus().equals(COMPLETE)) {
            return (logs == null);
        } else if (((PipelineCompletionEvent) event).getStatus().equals(EVENT_COMPLETION_FAILED)) {
            var log = logs.get(0);
            return (log.getStatus().equals(FAILURE) && (!((DeploymentLog) log).isDeploymentComplete()
                    || (((DeploymentLog) log).isDeploymentComplete()) && !((DeploymentLog) log).getArgoApplicationName().equals("") && !((DeploymentLog) log).getArgoRevisionHash().equals("")));
        }
        return false;
    }
}
