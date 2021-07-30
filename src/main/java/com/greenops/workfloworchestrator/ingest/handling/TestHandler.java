package com.greenops.workfloworchestrator.ingest.handling;

import com.greenops.workfloworchestrator.datamodel.event.Event;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.StepData;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.Test;

public interface TestHandler {

    void triggerTest(String pipelineRepoUrl, StepData stepData, boolean beforeTest, String gitCommitHash, Event event);
    void createAndRunTest(String stepName, String pipelineRepoUrl, Test test, int testNumber, String gitCommitHash, Event event);
}
