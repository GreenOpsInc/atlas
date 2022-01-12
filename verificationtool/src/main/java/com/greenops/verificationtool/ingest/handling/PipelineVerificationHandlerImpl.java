package com.greenops.verificationtool.ingest.handling;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.util.datamodel.event.Event;
import com.greenops.util.datamodel.event.TriggerStepEvent;
import com.greenops.util.datamodel.pipelinestatus.FailedStep;
import com.greenops.util.datamodel.pipelinestatus.PipelineStatus;
import com.greenops.verificationtool.datamodel.verification.DAG;
import com.greenops.verificationtool.datamodel.verification.Vertex;
import com.greenops.verificationtool.ingest.apiclient.workflowtrigger.WorkflowTriggerApi;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Component;

import java.util.List;

@Component
public class PipelineVerificationHandlerImpl implements PipelineVerificationHandler {
    private final String POPULATED = "POPULATED";
    private final WorkflowTriggerApi workflowTriggerApi;
    private final ObjectMapper objectMapper;

    @Autowired
    public PipelineVerificationHandlerImpl(WorkflowTriggerApi workflowTriggerApi, @Qualifier("objectMapper") ObjectMapper objectMapper) {
        this.workflowTriggerApi = workflowTriggerApi;
        this.objectMapper = objectMapper;
    }

    private PipelineStatus getPipelineStatus(Event event) {
        var pipelineStatusStr = this.workflowTriggerApi.getPipelineStatus(event.getOrgName(), event.getPipelineName(), event.getTeamName());
        System.out.println("Pipeline Status: " + pipelineStatusStr);
        try {
            return this.objectMapper.readValue(pipelineStatusStr, PipelineStatus.class);
        } catch (JsonProcessingException e) {
            throw new RuntimeException("Unable to serialize pipeline status into PipelineStatus Object " + e);
        }
    }

    @Override
    public Boolean verify(Event event, DAG dag) {
        var pipelineStatus = getPipelineStatus(event);
        var prevVertices = dag.getPreviousVertices(event);
        for (Vertex prevVertex : prevVertices) {
            if (event instanceof TriggerStepEvent
                    && !pipelineStatus.getProgressingSteps().contains(prevVertex.getStepName())
                    && !pipelineStatus.isCancelled()
                    && pipelineStatus.isStable()) {
                continue;
            } else if (pipelineStatus.getProgressingSteps().contains(prevVertex.getStepName())
                    && !pipelineStatus.isCancelled()
                    && pipelineStatus.isStable()) {
                continue;
            } else {
                return false;
            }
        }
        System.out.println(event.getOrgName() + " " + event.getPipelineName() + " " + event.getTeamName() + " " + event.getClass().getName() + " Pipeline Verification Passed!");
        return true;
    }

    @Override
    public Boolean verifyExpected(Event event, PipelineStatus expectedPipelineStatus) {
        var pipelineStatus = getPipelineStatus(event);
        for (String step : pipelineStatus.getProgressingSteps()) {
            if (!expectedPipelineStatus.getProgressingSteps().contains(step)) {
                return false;
            }
        }
        if (expectedPipelineStatus.isStable() != pipelineStatus.isStable()) return false;
        if (expectedPipelineStatus.isCancelled() != pipelineStatus.isCancelled()) return false;
        if (expectedPipelineStatus.getFailedSteps() == null && pipelineStatus.getFailedSteps() != null) return false;
        if (expectedPipelineStatus.getFailedSteps() != null && pipelineStatus.getFailedSteps() == null) return false;
        if (expectedPipelineStatus.getFailedSteps() == null && pipelineStatus.getFailedSteps() == null) return true;
        for (FailedStep failedStep : pipelineStatus.getFailedSteps()) {
            if (getFailedStep(failedStep.getStep(), expectedPipelineStatus.getFailedSteps()) == null) {
                return false;
            }
            FailedStep expectedFailedStep = getFailedStep(failedStep.getStep(), expectedPipelineStatus.getFailedSteps());
            if (expectedFailedStep == null) {
                return false;
            }
            if (expectedFailedStep.isDeploymentFailed() != failedStep.isDeploymentFailed()) {
                return false;
            }
            if (!expectedFailedStep.getBrokenTest().equals(failedStep.getBrokenTest()) &&
                    (expectedFailedStep.getBrokenTest().equals(this.POPULATED) && failedStep.getBrokenTest().equals(""))) {
                return false;
            }
            if (!expectedFailedStep.getBrokenTestLog().equals(failedStep.getBrokenTestLog()) &&
                    (expectedFailedStep.getBrokenTestLog().equals(this.POPULATED) && failedStep.getBrokenTestLog().equals(""))) {
                return false;
            }
        }
        return true;
    }

    private FailedStep getFailedStep(String stepName, List<FailedStep> failedSteps) {
        for (var failedStep : failedSteps) {
            if (failedStep.getStep().equals(stepName)) {
                return failedStep;
            }
        }
        return null;
    }
}
