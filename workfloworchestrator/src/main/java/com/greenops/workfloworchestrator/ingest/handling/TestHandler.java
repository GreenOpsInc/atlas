package com.greenops.workfloworchestrator.ingest.handling;

import com.greenops.util.datamodel.event.Event;
import com.greenops.util.datamodel.git.GitRepoSchemaInfo;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.StepData;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.Test;

public interface TestHandler {

    void triggerTest(GitRepoSchemaInfo gitRepoSchemaInfo, StepData stepData, boolean beforeTest, String gitCommitHash, Event event);
    void createAndRunTest(String clusterName, StepData stepData, GitRepoSchemaInfo gitRepoSchemaInfo, Test test, int testNumber, String gitCommitHash, Event event);
}
