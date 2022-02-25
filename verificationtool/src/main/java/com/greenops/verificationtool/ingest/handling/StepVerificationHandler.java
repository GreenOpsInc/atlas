package com.greenops.verificationtool.ingest.handling;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.greenops.util.datamodel.auditlog.Log;
import com.greenops.util.datamodel.event.Event;
import com.greenops.verificationtool.datamodel.verification.DAG;

import java.util.HashMap;
import java.util.List;

public interface StepVerificationHandler {
    Boolean verify(Event event, DAG dag);
    HashMap<String, String> verifyExpected(Event event, List<Log> logs) throws JsonProcessingException;
}
