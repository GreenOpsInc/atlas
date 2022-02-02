package com.greenops.workfloworchestrator.ingest.handling;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.util.datamodel.event.Event;
import com.greenops.util.datamodel.git.GitRepoSchemaInfo;
import com.greenops.util.datamodel.pipelinedata.StepData;
import com.greenops.util.datamodel.pipelinedata.Test;
import com.greenops.util.datamodel.request.GetFileRequest;
import com.greenops.workfloworchestrator.ingest.apiclient.clientwrapper.ClientRequestQueue;
import com.greenops.workfloworchestrator.ingest.apiclient.reposerver.RepoManagerApi;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Component;

import static com.greenops.workfloworchestrator.ingest.handling.DeploymentHandlerImpl.getStepNamespace;

@Slf4j
@Component
public class TestHandlerImpl implements TestHandler {

    private RepoManagerApi repoManagerApi;
    private ClientRequestQueue clientRequestQueue;
    private ObjectMapper yamlObjectMapper;

    @Autowired
    TestHandlerImpl(RepoManagerApi repoManagerApi, ClientRequestQueue clientRequestQueue, @Qualifier("yamlObjectMapper") ObjectMapper objectMapper) {
        this.repoManagerApi = repoManagerApi;
        this.clientRequestQueue = clientRequestQueue;
        this.yamlObjectMapper = objectMapper;
    }

    @Override
    public void triggerTest(GitRepoSchemaInfo gitRepoSchemaInfo, StepData stepData, boolean beforeTest, String gitCommitHash, Event event) {
        for (int i = 0; i < stepData.getTests().size(); i++) {
            if (beforeTest == stepData.getTests().get(i).shouldExecuteBefore()) {
                createAndRunTest(stepData.getClusterName(), stepData, gitRepoSchemaInfo, stepData.getTests().get(i), i, gitCommitHash, event);
                return;
            }
        }
    }

    @Override
    public void createAndRunTest(String clusterName, StepData stepData, GitRepoSchemaInfo gitRepoSchemaInfo, Test test, int testNumber, String gitCommitHash, Event event) {
        var getFileRequest = new GetFileRequest(gitRepoSchemaInfo, test.getPath(), gitCommitHash);
        var testConfig = repoManagerApi.getFileFromRepo(getFileRequest, event.getOrgName(), event.getTeamName());
        log.info("Creating test Job...");
        clientRequestQueue.deployAndWatch(
                clusterName,
                event.getOrgName(),
                event.getTeamName(),
                event.getPipelineName(),
                event.getPipelineUvn(),
                stepData.getName(),
                getStepNamespace(event, repoManagerApi, yamlObjectMapper, stepData.getArgoApplicationPath(), gitRepoSchemaInfo, gitCommitHash),
                ClientRequestQueue.DEPLOY_TEST_REQUEST,
                ClientRequestQueue.LATEST_REVISION,
                test.getPayload(testNumber, testConfig),
                test.getWatchKey(),
                testNumber
        );
    }
}
