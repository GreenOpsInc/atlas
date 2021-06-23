package com.greenops.workfloworchestrator.datamodel.pipelinedata;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;

@JsonDeserialize(as = PipelineDataImpl.class)
public interface PipelineData {
}
