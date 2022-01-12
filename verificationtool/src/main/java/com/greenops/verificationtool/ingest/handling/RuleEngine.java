package com.greenops.verificationtool.ingest.handling;

import com.greenops.util.datamodel.event.ApplicationInfraCompletionEvent;
import com.greenops.util.datamodel.event.Event;
import com.greenops.util.datamodel.event.PipelineCompletionEvent;
import com.greenops.util.datamodel.event.TriggerStepEvent;
import com.greenops.verificationtool.datamodel.rules.RuleData;
import org.springframework.stereotype.Component;

import java.util.HashMap;
import java.util.List;

@Component
public class RuleEngine {
    private final String PipelineTriggerEvent = "PipelineTriggerEvent";
    private final String TriggerStepEvent = "TriggerStepEvent";
    private final String ApplicationInfraCompletionEvent = "ApplicationInfraCompletionEvent";
    private final String ApplicationInfraTriggerEvent = "ApplicationInfraTriggerEvent";
    private final String ClientCompletionEvent = "ClientCompletionEvent";
    private final String TestCompletionEvent = "TestCompletionEvent";
    private final String PipelineCompletionEvent = "PipelineCompletionEvent";
    private final HashMap<String, List<RuleData>> rulesMapping;

    public RuleEngine() {
        this.rulesMapping = new HashMap<String, List<RuleData>>();
    }

    public void registerRules(String pipelineName, List<RuleData> rules) {
        rulesMapping.put(pipelineName, rules);
    }

    public RuleData getRule(Event event) {
        var pipelineName = event.getPipelineName();
        var stepName = event.getStepName();
        String eventType = null;
        if (event instanceof com.greenops.util.datamodel.event.PipelineTriggerEvent) {
            eventType = this.PipelineTriggerEvent;
        } else if (event instanceof com.greenops.util.datamodel.event.ClientCompletionEvent) {
            eventType = this.ClientCompletionEvent;
        } else if (event instanceof com.greenops.util.datamodel.event.TestCompletionEvent) {
            eventType = this.TestCompletionEvent;
        } else if (event instanceof com.greenops.util.datamodel.event.ApplicationInfraTriggerEvent) {
            eventType = this.ApplicationInfraTriggerEvent;
        } else if (event instanceof ApplicationInfraCompletionEvent) {
            eventType = this.ApplicationInfraCompletionEvent;
        } else if (event instanceof TriggerStepEvent) {
            eventType = this.TriggerStepEvent;
        } else if (event instanceof PipelineCompletionEvent) {
            eventType = this.PipelineCompletionEvent;
        }

        if (this.rulesMapping.get(pipelineName) == null) {
            return null;
        }
        var rules = this.rulesMapping.get(pipelineName);
        for (RuleData rule : rules) {
            if (stepName.equals(rule.getStepName()) && eventType.equals(rule.getEventType())) {
                return rule;
            }
        }
        return null;
    }
}
