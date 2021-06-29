package com.greenops.workfloworchestrator.ingest.handling;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.workfloworchestrator.datamodel.event.ClientCompletionEvent;
import com.greenops.workfloworchestrator.datamodel.event.Event;
import com.greenops.workfloworchestrator.datamodel.event.TestCompletionEvent;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.PipelineData;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.StepData;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.Test;
import com.greenops.workfloworchestrator.datamodel.pipelineschema.TeamSchema;
import com.greenops.workfloworchestrator.datamodel.requests.GetFileRequest;
import com.greenops.workfloworchestrator.datamodel.requests.KubernetesCreationRequest;
import com.greenops.workfloworchestrator.datamodel.requests.WatchRequest;
import com.greenops.workfloworchestrator.ingest.apiclient.clientwrapper.ClientWrapperApi;
import com.greenops.workfloworchestrator.ingest.apiclient.reposerver.RepoManagerApi;
import com.greenops.workfloworchestrator.ingest.dbclient.DbClient;
import com.greenops.workfloworchestrator.ingest.dbclient.DbKey;
import com.greenops.workfloworchestrator.ingest.handling.testautomation.CommandBuilder;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;

import static com.greenops.workfloworchestrator.ingest.handling.ClientKey.makeTestKey;

@Slf4j
@Component
public class EventHandlerImpl implements EventHandler {

    private static final String WATCH_ARGO_APPLICATION_KEY = "WatchArgoApplicationKey";
    private static final String WATCH_TEST_KEY = "WatchTestKey";
    private static final String PIPELINE_FILE_NAME = "pipeline.yaml";

    private ClientWrapperApi clientWrapperApi;
    private RepoManagerApi repoManagerApi;
    private DbClient dbClient;
    private CommandBuilder commandBuilder;
    private ObjectMapper yamlObjectMapper;
    private ObjectMapper objectMapper;

    @Autowired
    EventHandlerImpl(ClientWrapperApi clientWrapperApi,
                     RepoManagerApi repoManagerApi,
                     DbClient dbClient,
                     CommandBuilder commandBuilder,
                     @Qualifier("yamlObjectMapper") ObjectMapper yamlObjectMapper,
                     @Qualifier("objectMapper") ObjectMapper objectMapper) {
        this.clientWrapperApi = clientWrapperApi;
        this.repoManagerApi = repoManagerApi;
        this.dbClient = dbClient;
        this.commandBuilder = commandBuilder;
        this.yamlObjectMapper = yamlObjectMapper;
        this.objectMapper = objectMapper;
    }

    @Override
    public boolean handleEvent(Event event) {
        if (event instanceof ClientCompletionEvent) {
            return handleStepCompletion((ClientCompletionEvent) event);
        } else if (event instanceof TestCompletionEvent) {
            log.info("Test event received {}, {}", ((TestCompletionEvent) event).getTestName(), ((TestCompletionEvent) event).getSuccessful());
            //TODO: Ignore for now. Should trigger next tests
        }
        return true;
    }

    private boolean handleStepCompletion(ClientCompletionEvent event) {
        var teamSchema = fetchTeamSchema(event);
        if (teamSchema == null) return false;
        var pipelineData = fetchPipelineData(event, teamSchema);
        if (pipelineData == null) return false;
        var childrenSteps = pipelineData.getChildrenSteps(event.getStepName());
        for (var stepName : childrenSteps) {
            var step = pipelineData.getStep(stepName);
            if (areParentStepsComplete(stepName)) {
                var gitRepoUrl = teamSchema.getPipelineSchema(event.getPipelineName()).getGitRepoSchema().getGitRepo();
                if (!deployStep(event.getPipelineName(), gitRepoUrl, step, event)) return false;
            }
        }
        return true;
    }

    private boolean handleTestCompletion(TestCompletionEvent event) {
        var teamSchema = fetchTeamSchema(event);
        if (teamSchema == null) return false;
        var pipelineData = fetchPipelineData(event, teamSchema);
        if (pipelineData == null) return false;
        var step = pipelineData.getStep(event.getStepName());
        var gitRepoUrl = teamSchema.getPipelineSchema(event.getPipelineName()).getGitRepoSchema().getGitRepo();
        for (int i = 0; i < step.getTests().size(); i++) {
            var test = step.getTests().get(i);
            if (event.getTestName().equals(getFileName(test.getPath()))) {
                if ((test.shouldExecuteBefore() && i == step.getTests().size() - 1)
                        || (i < step.getTests().size() - 1 && test.shouldExecuteBefore() && !step.getTests().get(i + 1).shouldExecuteBefore())) {
                    //trigger current step deployment
                } else if (!test.shouldExecuteBefore() && i == step.getTests().size() - 1) {
                    //trigger next step
                } else if ((test.shouldExecuteBefore() && step.getTests().get(i + 1).shouldExecuteBefore())
                        || !test.shouldExecuteBefore()) {
                    return runStepTest(pipelineData.getName(), gitRepoUrl, step.getTests().get(i + 1), event);
                } else {
                    //This case should never be happening...log and see what the edge case is
                    log.info("EDGE CASE: {}, {}", test.shouldExecuteBefore(), i == step.getTests().size() - 1);
                }
            }
        }
        if (step.getTests().size() == 0) {
            //TODO: Should directly deploy step.
        }
        return false;
    }

    private boolean runStepTest(String pipelineName, String pipelineRepoUrl, Test test, Event event) {
        var getFileRequest = new GetFileRequest(pipelineRepoUrl, test.getPath());
        var testConfig = repoManagerApi.getFileFromRepo(getFileRequest, event.getOrgName(), event.getTeamName());
        var testKey = makeTestKey(event.getTeamName(), event.getPipelineName(), event.getStepName(), getFileNameWithoutExtension(test.getPath()));
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
            var watchRequest = new WatchRequest(event.getTeamName(), event.getPipelineName(), event.getStepName(), WATCH_TEST_KEY, testKey, deployResponse.getApplicationNamespace());
            var watching = clientWrapperApi.watchApplication(event.getOrgName(), watchRequest);
            if (!watching) return false;
            log.info("Watching Job");
        }
        return true;
    }

    private boolean deployStep(String pipelineName, String pipelineRepoUrl, StepData stepData, Event event) {
        //TODO: Test should not be run synchronously. This should be removed and events should be added to trigger before tests
        var beforeTests = stepData.getTests().stream().filter(Test::shouldExecuteBefore).collect(Collectors.toList());
        for (var test : beforeTests) {
            runStepTest(pipelineName, pipelineRepoUrl, test, event);
        }
        if (stepData.getOtherDeploymentsPath() != null) {
            var getFileRequest = new GetFileRequest(pipelineRepoUrl, stepData.getOtherDeploymentsPath());
            var otherDeploymentsConfig = repoManagerApi.getFileFromRepo(getFileRequest, event.getOrgName(), event.getTeamName());
            var deployResponse = clientWrapperApi.deploy(event.getOrgName(), ClientWrapperApi.DEPLOY_KUBERNETES_REQUEST, Optional.of(otherDeploymentsConfig), Optional.empty());
            if (!deployResponse.getSuccess()) {
                log.error("Deploying other resources failed.");
                return false;
            }
        }
        if (stepData.getArgoApplicationPath() != null) {
            var getFileRequest = new GetFileRequest(pipelineRepoUrl, stepData.getArgoApplicationPath());
            var argoApplicationConfig = repoManagerApi.getFileFromRepo(getFileRequest, event.getOrgName(), event.getTeamName());
            var deployResponse = clientWrapperApi.deploy(event.getOrgName(), ClientWrapperApi.DEPLOY_ARGO_REQUEST, Optional.of(argoApplicationConfig), Optional.empty());
            log.info("Deploying Argo application {}...", stepData.getArgoApplication());
            if (!deployResponse.getSuccess()) {
                log.error("Deploying the Argo application failed.");
                return false;
            } else {
                var watchRequest = new WatchRequest(event.getTeamName(), event.getPipelineName(), event.getStepName(), WATCH_ARGO_APPLICATION_KEY, stepData.getArgoApplication(), deployResponse.getApplicationNamespace());
                var watching = clientWrapperApi.watchApplication(event.getOrgName(), watchRequest);
                if (!watching) return false;
                log.info("Watching Argo application {}", stepData.getArgoApplication());
            }
        } else {
            //TODO: Expectation is argo application has already been created. Not currently supported.
        }
        //TODO: Audit log

        //TODO: Test should not be run synchronously. This should be removed and events should be added to trigger before tests
        var afterTests = stepData.getTests().stream().filter(test -> !test.shouldExecuteBefore()).collect(Collectors.toList());
        for (var test : afterTests) {
            runStepTest(pipelineName, pipelineRepoUrl, test, event);
        }
        return true;
    }

    public TeamSchema fetchTeamSchema(Event event) {
        return dbClient.fetchTeamSchema(DbKey.makeDbTeamKey(event.getOrgName(), event.getTeamName()));
    }

    private PipelineData fetchPipelineData(Event event, TeamSchema teamSchema) {
        var gitRepoUrl = teamSchema.getPipelineSchema(event.getPipelineName()).getGitRepoSchema().getGitRepo();
        var getFileRequest = new GetFileRequest(gitRepoUrl, PIPELINE_FILE_NAME);
        try {
            return objectMapper.readValue(
                    objectMapper.writeValueAsString(
                            yamlObjectMapper.readValue(repoManagerApi.getFileFromRepo(getFileRequest, event.getOrgName(), event.getTeamName()), Object.class)
                    ),
                    PipelineData.class);
        } catch (JsonProcessingException e) {
            log.error("Could not parse YAML pipeline data file", e);
        }
        return null;
    }

    private boolean areParentStepsComplete(String stepName) {
        //TODO: Check redis and implement the flow for ensuring the parent steps have all been completed
        return true;
    }

    //Reference for escaping file contents: https://stackoverflow.com/questions/15783701/which-characters-need-to-be-escaped-when-using-bash
    private String escapeFile(String fileContents) {
        var escapedFileContents = new StringBuilder();
        for (int i = 0; i < fileContents.length(); i++) {
            if (fileContents.charAt(i) == '\'') {
                escapedFileContents.append("'\\'");
            }
            escapedFileContents.append(fileContents.charAt(i));
        }
        return escapedFileContents.toString();
    }

    private String getFileName(String filePathAndName) {
        var splitPath = filePathAndName.split("/");
        var idx = splitPath.length - 1;
        while (idx >= 0) {
            if (splitPath[idx].equals("")) {
                idx--;
            } else {
                break;
            }
        }
        if (idx >= 0) {
            return splitPath[idx];
        }
        return null;
    }

    private String getFileNameWithoutExtension(String filePathAndName) {
        var filename = getFileName(filePathAndName);
        if (filename == null) return null;
        var idx = filename.length() - 1;
        while (idx >= 0) {
            if (filename.charAt(idx) == '.') {
                return filename.substring(0, idx);
            }
            idx--;
        }
        //Assuming its already an executable if there is no period
        return filename;
    }
}
