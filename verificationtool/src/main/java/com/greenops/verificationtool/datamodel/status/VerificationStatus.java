package com.greenops.verificationtool.datamodel.status;

import com.greenops.util.datamodel.event.Event;

import java.util.HashMap;

public interface VerificationStatus {
    public void markPipelineProgress(Event event);

    public void markPipelineComplete();

    public void markPipelineFailed(Event type, String failedType);

    public void markExpectedFailed(Event type, String failedType, HashMap<String, String> diff);
}
