package com.greenops.verificationtool.ingest.handling;

import com.greenops.util.datamodel.event.*;
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
    private final String FailureEvent = "FailureEvent";
    private final HashMap<String, List<RuleData>> rulesMapping;

    public RuleEngine() {
        this.rulesMapping = new HashMap<String, List<RuleData>>();
    }

    public void registerRules(String pipelineIdentifier, List<RuleData> rules) {
        rulesMapping.put(pipelineIdentifier, rules);
    }

    public RuleData getRule(String stepName, String pipelineName, String teamName, String eventType) {
        if (this.rulesMapping.get(pipelineName+ "-" + teamName) == null) {
            return null;
        }
        var rules = this.rulesMapping.get(pipelineName+ "-" + teamName);
        for (RuleData rule : rules) {
            if (stepName.equals(rule.getStepName()) && eventType.equals(rule.getEventType())) {
                return rule;
            }
        }
        return null;
    }
}
