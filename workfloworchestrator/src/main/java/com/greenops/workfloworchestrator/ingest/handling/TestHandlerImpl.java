package com.greenops.workfloworchestrator.ingest.handling;

import com.greenops.util.datamodel.event.Event;
import com.greenops.util.datamodel.request.GetFileRequest;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.StepData;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.Test;
import com.greenops.workfloworchestrator.ingest.apiclient.clientwrapper.ClientRequestQueue;
import com.greenops.workfloworchestrator.ingest.apiclient.reposerver.RepoManagerApi;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import static com.greenops.workfloworchestrator.ingest.handling.EventHandlerImpl.WATCH_TEST_KEY;

@Slf4j
@Component
public class TestHandlerImpl implements TestHandler {

    private RepoManagerApi repoManagerApi;
    private ClientRequestQueue clientRequestQueue;

    @Autowired
    TestHandlerImpl(RepoManagerApi repoManagerApi, ClientRequestQueue clientRequestQueue) {
        this.repoManagerApi = repoManagerApi;
        this.clientRequestQueue = clientRequestQueue;
    }

    @Override
    public void triggerTest(String pipelineRepoUrl, StepData stepData, boolean beforeTest, String gitCommitHash, Event event) {
        for (int i = 0; i < stepData.getTests().size(); i++) {
            if (beforeTest == stepData.getTests().get(i).shouldExecuteBefore()) {
                createAndRunTest(stepData.getClusterName(), stepData.getName(), pipelineRepoUrl, stepData.getTests().get(i), i, gitCommitHash, event);
                return;
            }
        }
    }

    @Override
    public void createAndRunTest(String clusterName, String stepName, String pipelineRepoUrl, Test test, int testNumber, String gitCommitHash, Event event) {
        var getFileRequest = new GetFileRequest(pipelineRepoUrl, test.getPath(), gitCommitHash);
        var testConfig = repoManagerApi.getFileFromRepo(getFileRequest, event.getOrgName(), event.getTeamName());
        log.info("Creating test Job...");
        clientRequestQueue.deployAndWatch(
                clusterName,
                event.getOrgName(),
                event.getTeamName(),
                event.getPipelineName(),
                event.getPipelineUvn(),
                stepName,
                ClientRequestQueue.DEPLOY_TEST_REQUEST,
                ClientRequestQueue.LATEST_REVISION,
                test.getPayload(testNumber, testConfig),
                test.getWatchKey(),
                testNumber
        );
    }
}
