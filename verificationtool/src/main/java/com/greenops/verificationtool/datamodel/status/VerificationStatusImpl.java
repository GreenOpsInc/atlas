package com.greenops.verificationtool.datamodel.status;

import com.greenops.util.datamodel.event.Event;
import com.greenops.util.datamodel.event.FailureEvent;

public class VerificationStatusImpl implements VerificationStatus {
    public static String IDLE = "IDLE";
    public static String PROGRESSING = "PROGRESSING";
    public static String FAILURE = "FAILURE";
    public static String COMPLETE = "COMPLETE";
    public static String EXPECTED_PIPELINE_STATUS_FAILED = "EXPECTED_PIPELINE_STATUS_FAILED";
    public static String EXPECTED_STEP_STATUS_FAILED = "EXPECTED_STEP_STATUS_FAILED";
    public static String EVENT_COMPLETION_FAILED = "EVENT_COMPLETION_FAILED";
    public static String FAILURE_EVENT_RECEIVED = "FAILURE_EVENT_RECEIVED";
    public static String EVENT_ORDER_VERIFICATION_FAILED = "EVENT_ORDER_VERIFICATION_FAILED";
    public static String PIPELINE_STATUS_VERIFICATION_FAILED = "PIPELINE_STATUS_VERIFICATION_FAILED";
    public static String STEP_STATUS_VERIFICATION_FAILED = "STEP_STATUS_VERIFICATION_FAILED";

    private String status;
    private String eventFailed;
    private String stepName;
    private String failedType;
    private String log;

    public VerificationStatusImpl() {
        this.status = IDLE;
        this.eventFailed = null;
        this.stepName = null;
        this.failedType = null;
        this.log = null;
    }

    @Override
    public void markPipelineProgress() {
        this.status = PROGRESSING;
    }

    @Override
    public void markPipelineComplete() {
        this.status = COMPLETE;
    }

    @Override
    public void markPipelineFailed(Event event, String failedType) {
        this.status = FAILURE;
        this.eventFailed = event.getClass().getName();
        this.failedType = failedType;
        this.stepName = event.getStepName();

        if (event instanceof FailureEvent) {
            this.log = ((FailureEvent) event).getError();
        }
    }
}
