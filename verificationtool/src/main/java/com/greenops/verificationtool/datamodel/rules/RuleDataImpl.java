package com.greenops.verificationtool.datamodel.rules;

import com.greenops.util.datamodel.auditlog.Log;
import com.greenops.util.datamodel.pipelinestatus.PipelineStatus;

public class RuleDataImpl implements RuleData {
    private final PipelineStatus pipelineStatus;
    private final Log log;
    private final String stepName;
    private final String eventType;

    public RuleDataImpl(String stepName, String eventType, PipelineStatus pipelineStatus, Log log) {
        this.stepName = stepName;
        this.eventType = eventType;
        this.pipelineStatus = pipelineStatus;
        this.log = log;
    }

    @Override
    public String getStepName() {
        return stepName;
    }

    @Override
    public String getEventType() {
        return eventType;
    }

    @Override
    public PipelineStatus getPipelineStatus() {
        return pipelineStatus;
    }

    @Override
    public Log getStepStatus() {
        return log;
    }
}
