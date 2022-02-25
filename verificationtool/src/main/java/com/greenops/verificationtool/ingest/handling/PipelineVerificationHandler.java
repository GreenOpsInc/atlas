package com.greenops.verificationtool.ingest.handling;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.greenops.util.datamodel.event.Event;
import com.greenops.util.datamodel.pipelinestatus.PipelineStatus;
import com.greenops.verificationtool.datamodel.verification.DAG;

import java.util.HashMap;

public interface PipelineVerificationHandler {
    Boolean verify(Event event, DAG dag);

    HashMap<String, String> verifyExpected(Event event, PipelineStatus expectedPipelineStatus) throws JsonProcessingException;
}
