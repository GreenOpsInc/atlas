package com.greenops.verificationtool.ingest.handling;

import com.greenops.util.datamodel.event.Event;
import com.greenops.util.datamodel.pipelinestatus.PipelineStatus;
import com.greenops.verificationtool.datamodel.verification.DAG;

public interface PipelineVerificationHandler {
    Boolean verify(Event event, DAG dag);

    Boolean verifyExpected(Event event, PipelineStatus expectedPipelineStatus);
}
