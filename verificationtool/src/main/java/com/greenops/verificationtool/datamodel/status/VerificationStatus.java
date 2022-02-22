package com.greenops.verificationtool.datamodel.status;

import com.greenops.util.datamodel.event.Event;

public interface VerificationStatus {
    public void markPipelineProgress(Event event);

    public void markPipelineComplete();

    public void markPipelineFailed(Event type, String failedType);
}
