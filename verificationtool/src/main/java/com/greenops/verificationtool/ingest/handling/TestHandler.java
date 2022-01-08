package com.greenops.verificationtool.ingest.handling;

import com.greenops.util.datamodel.event.Event;
import com.greenops.verificationtool.datamodel.pipelinedata.StepData;
import com.greenops.verificationtool.datamodel.pipelinedata.Test;

public interface TestHandler {

    void triggerTest(String pipelineRepoUrl, StepData stepData, boolean beforeTest, String gitCommitHash, Event event);
    void createAndRunTest(String clusterName, StepData stepData, String pipelineRepoUrl, Test test, int testNumber, String gitCommitHash, Event event);
}
