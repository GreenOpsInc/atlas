package com.greenops.workfloworchestrator.ingest.handling;

import com.greenops.workfloworchestrator.datamodel.event.Event;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.StepData;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.Test;

public interface TestHandler {

    boolean triggerTest(String pipelineRepoUrl, StepData stepData, boolean beforeTest, Event event);
    boolean createAndRunTest(String stepName, String pipelineRepoUrl, Test test, int testNumber, Event event);
}
