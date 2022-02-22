package com.greenops.verificationtool.datamodel.rules;

import com.greenops.util.datamodel.auditlog.Log;
import com.greenops.util.datamodel.pipelinestatus.PipelineStatus;

import java.util.List;

public class RuleDataImpl implements RuleData {
    private final PipelineStatus pipelineStatus;
    private final List<Log> logs;
    private final String stepName;
    private final String eventType;

    public RuleDataImpl(String stepName, String eventType, PipelineStatus pipelineStatus, List<Log> logs) {
        this.stepName = stepName;
        this.eventType = eventType;
        this.pipelineStatus = pipelineStatus;
        this.logs = logs;
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
    public List<Log> getStepStatus(){
        return logs;
    }
}
