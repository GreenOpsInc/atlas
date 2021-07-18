package com.greenops.workfloworchestrator.ingest.handling;

import com.greenops.workfloworchestrator.datamodel.event.Event;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.StepData;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.Test;
import com.greenops.workfloworchestrator.datamodel.requests.GetFileRequest;
import com.greenops.workfloworchestrator.datamodel.requests.KubernetesCreationRequest;
import com.greenops.workfloworchestrator.datamodel.requests.WatchRequest;
import com.greenops.workfloworchestrator.ingest.apiclient.clientwrapper.ClientWrapperApi;
import com.greenops.workfloworchestrator.ingest.apiclient.reposerver.RepoManagerApi;
import com.greenops.workfloworchestrator.ingest.handling.testautomation.CommandBuilder;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.Optional;

import static com.greenops.workfloworchestrator.ingest.handling.ClientKey.makeTestKey;
import static com.greenops.workfloworchestrator.ingest.handling.EventHandlerImpl.WATCH_TEST_KEY;
import static com.greenops.workfloworchestrator.ingest.handling.util.deployment.SchemaHandlingUtil.escapeFile;
import static com.greenops.workfloworchestrator.ingest.handling.util.deployment.SchemaHandlingUtil.getFileName;

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
    public boolean triggerTest(String pipelineRepoUrl, StepData stepData, boolean beforeTest, Event event) {
        for (int i = 0; i < stepData.getTests().size(); i++) {
            if (beforeTest == stepData.getTests().get(i).shouldExecuteBefore()) {
                return createAndRunTest(stepData.getName(), pipelineRepoUrl, stepData.getTests().get(i), i, event);
            }
        }
        return true;
    }

    @Override
    public boolean createAndRunTest(String stepName, String pipelineRepoUrl, Test test, int testNumber, Event event) {
        var getFileRequest = new GetFileRequest(pipelineRepoUrl, test.getPath());
        var testConfig = repoManagerApi.getFileFromRepo(getFileRequest, event.getOrgName(), event.getTeamName());
        var testKey = makeTestKey(testNumber);
        var filename = getFileName(test.getPath());
        var creationRequest = new KubernetesCreationRequest(
                "Job",
                testKey,
                "",
                "",
                List.of("/bin/sh", "-c"),
                new CommandBuilder().createFile(filename, escapeFile(testConfig)).compile(filename).executeExistingFile(filename).build(),
                test.getVariables()
        );
        log.info("Creating test Job...");
        var deployResponse = clientWrapperApi.deploy(event.getOrgName(), ClientWrapperApi.DEPLOY_TEST_REQUEST, Optional.empty(), Optional.of(creationRequest));
        if (deployResponse.getSuccess()) {
            var watchRequest = new WatchRequest(event.getTeamName(), event.getPipelineName(), stepName, WATCH_TEST_KEY, testKey, deployResponse.getApplicationNamespace());
            var watching = clientWrapperApi.watchApplication(event.getOrgName(), watchRequest);
            if (!watching) return false;
            log.info("Watching Job");
            return true;
        } else {
            return false;
        }
    }
}
