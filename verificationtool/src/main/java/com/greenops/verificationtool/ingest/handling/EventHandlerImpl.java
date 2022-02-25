package com.greenops.verificationtool.ingest.handling;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.greenops.util.datamodel.event.*;
import com.greenops.verificationtool.datamodel.verification.DAG;
import com.greenops.verificationtool.datamodel.verification.Vertex;
import com.greenops.verificationtool.ingest.kafka.KafkaClient;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.util.concurrent.TimeUnit;

import static com.greenops.verificationtool.datamodel.status.VerificationStatusImpl.*;

@Slf4j
@Component
public class EventHandlerImpl implements EventHandler {
    private final String HEALTHY = "Healthy";
    private final String ATLAS_ROOT_DATA = "ATLAS_ROOT_DATA";
    private final String PipelineCompletionEvent = "PipelineCompletionEvent";
    private final String ApplicationInfraCompletionEvent = "ApplicationInfraCompletionEvent";
    private final String ClientCompletionEvent = "ClientCompletionEvent";
    private final String TestCompletionEvent = "TestCompletionEvent";
    private final String FailureEvent = "FailureEvent";
    private final String verificationTopicName;
    private final KafkaClient kafkaClient;
    private final DagRegistry dagRegistry;
    private final VerificationStatusRegistry verificationStatusRegistry;
    private final RuleEngine ruleEngine;
    private final PipelineVerificationHandler pipelineVerificationHandler;
    private final StepVerificationHandler stepVerificationHandler;

    @Autowired
    EventHandlerImpl(KafkaClient kafkaClient,
                     DagRegistry dagRegistry,
                     VerificationStatusRegistry verificationStatusRegistry,
                     RuleEngine ruleEngine,
                     PipelineVerificationHandler pipelineVerificationHandler,
                     StepVerificationHandler stepVerificationHandler,
                     @Value("${spring.kafka.verification-topic}") String verificationTopicName) {
        this.kafkaClient = kafkaClient;
        this.dagRegistry = dagRegistry;
        this.verificationStatusRegistry = verificationStatusRegistry;
        this.ruleEngine = ruleEngine;
        this.pipelineVerificationHandler = pipelineVerificationHandler;
        this.stepVerificationHandler = stepVerificationHandler;
        this.verificationTopicName = verificationTopicName;
    }

    @Override
    public void handleEvent(Event event) throws InterruptedException {
        log.info("____________________________________");
        log.info("Handling event of type {}", event.getClass().getName());
        var verificationStatus = this.verificationStatusRegistry.getVerificationStatus(event.getPipelineName() + "-" + event.getTeamName());
        var dag = dagRegistry.retrieveDagObj(event.getPipelineName() + "-" + event.getTeamName());
        try {
            handleExpectedRules(event, dag);
        } catch (JsonProcessingException e) {
            throw new RuntimeException("Unable to serialize status object into status string " + e);
        }

        if (!(event instanceof PipelineCompletionEvent)) {
            verificationStatus.markPipelineProgress(event);
        }
        if (isFailedEvent(event)) {
            return;
        }

        if (dagRegistry.retrieveDagObj(event.getPipelineName() + "-" + event.getTeamName()) == null) {
            throw new RuntimeException("Pipeline Does not exists in Dag Registry");
        }
        if (event instanceof PipelineTriggerEvent) {
            if (verifyEventOrder(event, dag)) {
                log.info("Verification passed for {}-{}-{}", event.getTeamName(), event.getStepName(), event.getPipelineName());
            } else {
                log.info("Verification failed for {}-{}-{}", event.getTeamName(), event.getStepName(), event.getPipelineName());
            }
        } else if (event instanceof PipelineCompletionEvent) {
            if (verifyPipelineStatus(event, dag) && verifyStepStatus(event, dag)) {
                log.info("Verification passed for {}-{}-{}", event.getTeamName(), event.getStepName(), event.getPipelineName());
            } else {
                log.info("Verification failed for {}-{}-{}", event.getTeamName(), event.getStepName(), event.getPipelineName());
            }
        } else {
            if (verifyEventOrder(event, dag) && verifyPipelineStatus(event, dag) && verifyStepStatus(event, dag)) {
                log.info("Verification passed for {}-{}-{}", event.getTeamName(), event.getStepName(), event.getPipelineName());
                if (dag.isAllVerticesVisited()) {
                    verificationStatus.markPipelineComplete();
                    sendCompletionEvent(event, COMPLETE);
                }
            } else {
                log.info("Verification failed for {}-{}-{}", event.getTeamName(), event.getStepName(), event.getPipelineName());
            }
        }
    }

    private void handleExpectedRules(Event event, DAG dag) throws JsonProcessingException {
        var verificationStatus = this.verificationStatusRegistry.getVerificationStatus(event.getPipelineName() + "-" + event.getTeamName());
        String stepName, pipelineName, teamName, eventType = null;
        stepName = event.getStepName();
        pipelineName = event.getPipelineName();
        teamName = event.getTeamName();
        if (event instanceof PipelineCompletionEvent) {
            eventType = ((PipelineCompletionEvent) event).getFailedEvent();
        } else {
            var parents = dag.getPreviousVertices(event);
            if (parents != null) {
                for (Vertex parent : parents) {
                    if (parent.getStepName().equals(event.getStepName())) {
                        eventType = parent.getEventType();
                        if(parent.getEventType().equals(TestCompletionEvent)){
                            eventType += parent.getTestNumber();
                        }
                        break;
                    }
                }
            }
        }
        if (eventType == null) return;

        var ruleData = this.ruleEngine.getRule(stepName, pipelineName, teamName, eventType);
        if (ruleData != null) {
            if (ruleData.getPipelineStatus() != null) {
                var expectedDiff = this.pipelineVerificationHandler.verifyExpected(event, ruleData.getPipelineStatus());
                if (expectedDiff == null) {
                    log.info("Expected Pipeline Status passed for {}-{}-{}", teamName, stepName, pipelineName);
                } else {
                    log.info("Expected Pipeline Status failed for {}-{}-{}", teamName, stepName, pipelineName);
                    verificationStatus.markExpectedFailed(event, EXPECTED_PIPELINE_STATUS_FAILED, expectedDiff);
                }
            }
            if (ruleData.getStepStatus() != null) {
                var expectedDiff = this.stepVerificationHandler.verifyExpected(event, ruleData.getStepStatus());
                if (expectedDiff == null) {
                    log.info("Expected Step Status passed for {}-{}-{}", teamName, stepName, pipelineName);
                } else {
                    log.info("Expected Step Status failed for {}-{}-{}", teamName, stepName, pipelineName);
                    verificationStatus.markExpectedFailed(event, EXPECTED_STEP_STATUS_FAILED, expectedDiff);
                }
            }
        }
    }

    private Boolean verifyEventOrder(Event event, DAG dag) {
        var verificationStatus = this.verificationStatusRegistry.getVerificationStatus(event.getPipelineName() + "-" + event.getTeamName());
        if (dag.checkEventOrderInDag(event)) {
            log.info("Event order verification passed for {}-{}-{}", event.getTeamName(), event.getStepName(), event.getPipelineName());
            return true;
        } else {
            log.info("Event order verification failed for {}-{}-{}", event.getTeamName(), event.getStepName(), event.getPipelineName());
            verificationStatus.markPipelineFailed(event, EVENT_ORDER_VERIFICATION_FAILED);
            return false;
        }
    }

    private Boolean verifyPipelineStatus(Event event, DAG dag) {
        var verificationStatus = this.verificationStatusRegistry.getVerificationStatus(event.getPipelineName() + "-" + event.getTeamName());
        if (this.pipelineVerificationHandler.verify(event, dag)) {
            log.info("Pipeline status verification passed for {}-{}-{}", event.getTeamName(), event.getStepName(), event.getPipelineName());
            return true;
        } else {
            log.info("Pipeline status verification failed for {}-{}-{}", event.getTeamName(), event.getStepName(), event.getPipelineName());
            verificationStatus.markPipelineFailed(event, PIPELINE_STATUS_VERIFICATION_FAILED);
            return false;
        }
    }

    private Boolean verifyStepStatus(Event event, DAG dag) {
        var verificationStatus = this.verificationStatusRegistry.getVerificationStatus(event.getPipelineName() + "-" + event.getTeamName());
        if (this.stepVerificationHandler.verify(event, dag)) {
            log.info("Step status verification passed for {}-{}-{}", event.getTeamName(), event.getStepName(), event.getPipelineName());
            return true;
        } else {
            log.info("Step status verification failed for {}-{}-{}", event.getTeamName(), event.getStepName(), event.getPipelineName());
            verificationStatus.markPipelineFailed(event, STEP_STATUS_VERIFICATION_FAILED);
            return false;
        }
    }

    private Boolean isFailedEvent(Event event) throws InterruptedException {
        var verificationStatus = this.verificationStatusRegistry.getVerificationStatus(event.getPipelineName() + "-" + event.getTeamName());
        if ((event instanceof ApplicationInfraCompletionEvent && !((ApplicationInfraCompletionEvent) event).isSuccess()) ||
                (event instanceof ClientCompletionEvent && !((ClientCompletionEvent) event).getHealthStatus().equals(this.HEALTHY)) ||
                (event instanceof TestCompletionEvent && !((TestCompletionEvent) event).getSuccessful())) {
            verificationStatus.markPipelineFailed(event, EVENT_COMPLETION_FAILED);
            sendCompletionEvent(event, EVENT_COMPLETION_FAILED);
            return true;
        } else if (event instanceof FailureEvent) {
            verificationStatus.markPipelineFailed(event, FAILURE_EVENT_RECEIVED);
            sendCompletionEvent(event, FAILURE_EVENT_RECEIVED);
            return true;
        }
        return false;
    }

    private void sendCompletionEvent(Event event, String status) throws InterruptedException {
        TimeUnit.SECONDS.sleep(1);
        var eventFailed = this.PipelineCompletionEvent;
        if (!status.equals(COMPLETE)) {
            if (event instanceof TestCompletionEvent) {
                eventFailed = this.TestCompletionEvent + ((TestCompletionEvent) event).getTestNumber();
            }
            else if (event instanceof ClientCompletionEvent) eventFailed = this.ClientCompletionEvent;
            else if (event instanceof ApplicationInfraCompletionEvent)
                eventFailed = this.ApplicationInfraCompletionEvent;
            else if (event instanceof FailureEvent) eventFailed = this.FailureEvent;
        }
        PipelineCompletionEvent pipelineCompletionEvent = new PipelineCompletionEvent(event.getOrgName(),
                event.getTeamName(),
                event.getPipelineName(),
                status.equals(COMPLETE) ? ATLAS_ROOT_DATA : event.getStepName(),
                event.getPipelineUvn(),
                status,
                eventFailed);
        this.kafkaClient.sendMessage(pipelineCompletionEvent, this.verificationTopicName);
    }
}