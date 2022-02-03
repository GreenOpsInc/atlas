package com.greenops.verificationtool.ingest.handling;

import com.greenops.util.datamodel.event.*;
import com.greenops.verificationtool.datamodel.verification.DAG;
import com.greenops.verificationtool.ingest.kafka.KafkaClient;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.util.concurrent.TimeUnit;

@Slf4j
@Component
public class EventHandlerImpl implements EventHandler {
    private final String HEALTHY = "Healthy";
    private final KafkaClient kafkaClient;
    private final DagRegistry dagRegistry;
    private final RuleEngine ruleEngine;
    private final PipelineVerificationHandler pipelineVerificationHandler;
    private final StepVerificationHandler stepVerificationHandler;

    @Autowired
    EventHandlerImpl(KafkaClient kafkaClient,
                     DagRegistry dagRegistry,
                     RuleEngine ruleEngine,
                     PipelineVerificationHandler pipelineVerificationHandler,
                     StepVerificationHandler stepVerificationHandler) {
        this.kafkaClient = kafkaClient;
        this.dagRegistry = dagRegistry;
        this.ruleEngine = ruleEngine;
        this.pipelineVerificationHandler = pipelineVerificationHandler;
        this.stepVerificationHandler = stepVerificationHandler;
    }

    @Override
    public void handleEvent(Event event) {
        log.info("Handling event of type {}", event.getClass().getName());
        var ruleData = this.ruleEngine.getRule(event);
        if (ruleData != null) {
            if (ruleData.getPipelineStatus() != null) {
                if (this.pipelineVerificationHandler.verifyExpected(event, ruleData.getPipelineStatus())) {
                    System.out.println("Expected Pipeline Status Passed for Event" + event.getClass().getName() + " " + event.getStepName() + " " + event.getPipelineName());
                } else {
                    System.out.println("Expected Pipeline Status Failed");
                }
            }
            if (ruleData.getStepStatus() != null) {
                if (this.stepVerificationHandler.verifyExpected(event, ruleData.getStepStatus())) {
                    System.out.println("Expected Step Status Passed for Event" + event.getClass().getName() + " " + event.getStepName() + " " + event.getPipelineName());
                } else {
                    System.out.println("Expected Step Status Failed");
                }
            }

        }

        this.dagRegistry.markPipelineProgress(event);
        if (dagRegistry.retriveDagObj(event.getPipelineName()) == null) {
            throw new RuntimeException("Pipeline Does not exists in Dag Registry");
        }
        var dag = dagRegistry.retriveDagObj(event.getPipelineName());
        if (event instanceof PipelineTriggerEvent) {
            if (dag.checkEventOrderInDag(event)) {
                System.out.println(event.getClass().getName() + " Verification Passed!");
            } else {
                System.out.println(event.getClass().getName() + " Verification Failed!");
            }
            return;
        }
        try {
            if (dag.checkEventOrderInDag(event) && this.pipelineVerificationHandler.verify(event, dag) && this.stepVerificationHandler.verify(event, dag)) {
                System.out.println(event.getClass().getName() + " Verification Passed!");
                if (this.isLastEvent(event, dag)) {
                    TimeUnit.SECONDS.sleep(1);
                    PipelineCompletionEvent pipelineCompletionEvent = new PipelineCompletionEvent(event.getOrgName(),
                            event.getTeamName(),
                            event.getPipelineName(),
                            event.getPipelineUvn());
                    this.kafkaClient.sendMessage(pipelineCompletionEvent);
                }
            } else {
                System.out.println(event.getClass().getName() + " Failed!");
            }
        } catch (InterruptedException e) {
            e.printStackTrace();
        }
    }

    private Boolean isLastEvent(Event event, DAG dag) {
        if (event instanceof ApplicationInfraCompletionEvent && !((ApplicationInfraCompletionEvent) event).isSuccess())
            return true;
        else if (event instanceof ClientCompletionEvent && !((ClientCompletionEvent) event).getHealthStatus().equals(this.HEALTHY)) {
            return true;
        } else if (event instanceof TestCompletionEvent && !((TestCompletionEvent) event).getSuccessful()) return true;
        else if (event instanceof FailureEvent) return true;
        else return dag.isLastVertexInDAG(event);
    }
}