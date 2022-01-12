package com.greenops.verificationtool.ingest.handling;

import com.greenops.util.datamodel.auditlog.Log;
import com.greenops.util.datamodel.event.Event;
import com.greenops.verificationtool.datamodel.verification.DAG;

public interface StepVerificationHandler {
    Boolean verify(Event event, DAG dag);

    Boolean verifyExpected(Event event, Log log);
}
