package com.greenops.verificationtool.datamodel.rules;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import com.greenops.util.datamodel.auditlog.Log;
import com.greenops.util.datamodel.pipelinestatus.PipelineStatus;

@JsonDeserialize(as = RuleDataImpl.class)
public interface RuleData {
    String getStepName();

    String getEventType();

    PipelineStatus getPipelineStatus();

    Log getStepStatus();
}
