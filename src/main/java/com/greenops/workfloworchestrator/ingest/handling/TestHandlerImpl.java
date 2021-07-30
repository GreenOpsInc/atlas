package com.greenops.workfloworchestrator.ingest.handling;

import com.greenops.workfloworchestrator.datamodel.event.Event;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.StepData;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.Test;
import com.greenops.workfloworchestrator.datamodel.requests.GetFileRequest;
import com.greenops.workfloworchestrator.datamodel.requests.WatchRequest;
import com.greenops.workfloworchestrator.error.AtlasRetryableError;
import com.greenops.workfloworchestrator.ingest.apiclient.clientwrapper.ClientWrapperApi;
import com.greenops.workfloworchestrator.ingest.apiclient.reposerver.RepoManagerApi;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import static com.greenops.workfloworchestrator.ingest.handling.EventHandlerImpl.WATCH_TEST_KEY;

@Slf4j
@Component
public class TestHandlerImpl implements TestHandler {

    private RepoManagerApi repoManagerApi;
    private ClientWrapperApi clientWrapperApi;

    @Autowired
    TestHandlerImpl(RepoManagerApi repoManagerApi, ClientWrapperApi clientWrapperApi) {
        this.repoManagerApi = repoManagerApi;
        this.clientWrapperApi = clientWrapperApi;
    }

    @Override
    public void triggerTest(String pipelineRepoUrl, StepData stepData, boolean beforeTest, String gitCommitHash, Event event) {
        for (int i = 0; i < stepData.getTests().size(); i++) {
            if (beforeTest == stepData.getTests().get(i).shouldExecuteBefore()) {
                createAndRunTest(stepData.getName(), pipelineRepoUrl, stepData.getTests().get(i), i, gitCommitHash, event);
                return;
            }
        }
    }

    @Override
    public void createAndRunTest(String stepName, String pipelineRepoUrl, Test test, int testNumber, String gitCommitHash, Event event) {
        var getFileRequest = new GetFileRequest(pipelineRepoUrl, test.getPath(), gitCommitHash);
        var testConfig = repoManagerApi.getFileFromRepo(getFileRequest, event.getOrgName(), event.getTeamName());
        log.info("Creating test Job...");
        var deployResponse = clientWrapperApi.deploy(
                event.getOrgName(),
                ClientWrapperApi.DEPLOY_TEST_REQUEST,
                test.getPayload(testNumber, testConfig)
        );
        if (deployResponse.getSuccess()) {
            var watchRequest = new WatchRequest(event.getTeamName(), event.getPipelineName(), stepName, WATCH_TEST_KEY, deployResponse.getResourceName(), deployResponse.getApplicationNamespace(), testNumber);
            clientWrapperApi.watchApplication(event.getOrgName(), watchRequest);
            log.info("Watching Job");
        } else {
            throw new AtlasRetryableError("Test deployment was unsuccessful");
        }
    }
}
