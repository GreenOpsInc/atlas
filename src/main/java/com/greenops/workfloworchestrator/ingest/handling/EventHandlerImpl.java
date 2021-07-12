package com.greenops.workfloworchestrator.ingest.handling;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.workfloworchestrator.datamodel.auditlog.DeploymentLog;
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

import static com.greenops.workfloworchestrator.datamodel.pipelinedata.StepData.ROOT_STEP_NAME;
import static com.greenops.workfloworchestrator.datamodel.pipelinedata.StepData.createRootStep;
import static com.greenops.workfloworchestrator.ingest.handling.ClientKey.getTestNumberFromTestKey;
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
        var teamSchema = fetchTeamSchema(event);
        if (teamSchema == null) return false;
        var pipelineData = fetchPipelineData(event, teamSchema);
        if (pipelineData == null) return false;
        var gitRepoUrl = teamSchema.getPipelineSchema(event.getPipelineName()).getGitRepoSchema().getGitRepo();
        if (event instanceof ClientCompletionEvent) {
            return handleClientCompletionEvent(pipelineData, gitRepoUrl, (ClientCompletionEvent) event);
        } else if (event instanceof TestCompletionEvent) {
            return handleTestCompletion(pipelineData, gitRepoUrl, (TestCompletionEvent) event);
        }
        return true;
    }

    private boolean handleClientCompletionEvent(PipelineData pipelineData, String pipelineRepoUrl, ClientCompletionEvent event) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), event.getStepName());
        var deploymentLog = dbClient.fetchLatestLog(logKey);
        if (deploymentLog != null) {
            deploymentLog.setDeploymentComplete(true);
            if (!dbClient.updateHeadInList(logKey, deploymentLog)) return false;
        }
        if (event.getStepName().equals(ROOT_STEP_NAME)) {
            return triggerNextSteps(pipelineData, createRootStep(), pipelineRepoUrl, event);
        }
        var step = pipelineData.getStep(event.getStepName());
        var afterTestsExist = step.getTests().stream().anyMatch(test -> !test.shouldExecuteBefore());
        if (afterTestsExist) {
            return triggerTest(pipelineRepoUrl, step, false, event);
        } else {
            return triggerNextSteps(pipelineData, step, pipelineRepoUrl, event);
        }
    }

    private boolean handleTestCompletion(PipelineData pipelineData, String pipelineRepoUrl, TestCompletionEvent event) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), event.getStepName());
        var deploymentLog = dbClient.fetchLatestLog(logKey);
        if (deploymentLog != null && !event.getSuccessful()) {
            deploymentLog.setBrokenTest(event.getTestName());
            deploymentLog.setBrokenTestLog(event.getLog());
            deploymentLog.setStatus(DeploymentLog.DeploymentStatus.FAILURE.name());
            //If the dbClient operation fails, return that, otherwise the pipeline should stop given the step failure
            return dbClient.updateHeadInList(logKey, deploymentLog);
        }

        var step = pipelineData.getStep(event.getStepName());
        var completedTestNumber = getTestNumberFromTestKey(event.getTestName());
        if (completedTestNumber < 0 || step.getTests().size() <= completedTestNumber) {
            log.info("Malformed test key or tests have changed");
            return false;
        }
        var completedTest = step.getTests().get(completedTestNumber);
        var tests = step.getTests().stream().filter(test -> test.shouldExecuteBefore() == completedTest.shouldExecuteBefore()).collect(Collectors.toList());

        if (completedTest.shouldExecuteBefore() && completedTestNumber == tests.size() - 1) {
            return deploy(event.getPipelineName(), pipelineRepoUrl, step, event);
        } else if (!completedTest.shouldExecuteBefore() && completedTestNumber == tests.size() - 1) {
            return triggerNextSteps(pipelineData, step, pipelineRepoUrl, event);
        } else if (completedTestNumber < tests.size()) {
            return runStepTest(step.getName(), pipelineRepoUrl, step.getTests().get(completedTestNumber + 1), completedTestNumber + 1, event);
        } else {
            //This case should never be happening...log and see what the edge case is
            log.info("EDGE CASE: {}, {}", completedTest.shouldExecuteBefore(), completedTestNumber == step.getTests().size() - 1);
        }
        return true;
    }

    private boolean runStepTest(String stepName, String pipelineRepoUrl, Test test, int testNumber, Event event) {
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

    private boolean triggerNextSteps(PipelineData pipelineData, StepData step, String pipelineRepoUrl, Event event) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), event.getStepName());
        var deploymentLog = dbClient.fetchLatestLog(logKey);
        if (deploymentLog != null && deploymentLog.getBrokenTest() == null) {
            deploymentLog.setStatus(DeploymentLog.DeploymentStatus.SUCCESS.name());
            if (!dbClient.updateHeadInList(logKey, deploymentLog)) return false;
        }

        var childrenSteps = pipelineData.getChildrenSteps(step.getName());
        for (var stepName : childrenSteps) {
            var nextStep = pipelineData.getStep(stepName);
            if (areParentStepsComplete(stepName)) {
                if (!triggerStep(event.getPipelineName(), pipelineRepoUrl, nextStep, event)) return false;
            }
        }
        return true;
    }

    private boolean triggerStep(String pipelineName, String pipelineRepoUrl, StepData stepData, Event event) {
        var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepData.getName());
        var newLog = new DeploymentLog(DeploymentLog.DeploymentStatus.PROGRESSING.name(), false, "_____");
        if (!dbClient.insertValueInList(logKey, newLog)) return false;
        var beforeTestsExist = stepData.getTests().stream().anyMatch(Test::shouldExecuteBefore);
        if (beforeTestsExist) {
            return triggerTest(pipelineRepoUrl, stepData, true, event);
        }

        if (stepData.getOtherDeploymentsPath() != null || stepData.getArgoApplicationPath() != null) {
            return deploy(pipelineName, pipelineRepoUrl, stepData, event);
        }

        var afterTestsExist = stepData.getTests().stream().anyMatch(test -> !test.shouldExecuteBefore());
        if (afterTestsExist) {
            return triggerTest(pipelineRepoUrl, stepData, false, event);
        }
        return true;
    }

    private boolean triggerTest(String pipelineRepoUrl, StepData stepData, boolean beforeTest, Event event) {
        for (int i = 0; i < stepData.getTests().size(); i++) {
            if (beforeTest == stepData.getTests().get(i).shouldExecuteBefore()) {
                return runStepTest(stepData.getName(), pipelineRepoUrl, stepData.getTests().get(i), i, event);
            }
        }
        return true;
    }

    private boolean deploy(String pipelineName, String pipelineRepoUrl, StepData stepData, Event event) {
        if (stepData.getOtherDeploymentsPath() != null) {
            var getFileRequest = new GetFileRequest(pipelineRepoUrl, stepData.getOtherDeploymentsPath());
            var otherDeploymentsConfig = repoManagerApi.getFileFromRepo(getFileRequest, event.getOrgName(), event.getTeamName());
            //TODO: The splitting of the config file should eventually be done on the client side
            for (var deploymentConfig : otherDeploymentsConfig.split("---")) {
                var deployResponse = clientWrapperApi.deploy(event.getOrgName(), ClientWrapperApi.DEPLOY_KUBERNETES_REQUEST, Optional.of(deploymentConfig), Optional.empty());
                if (!deployResponse.getSuccess()) {
                    log.error("Deploying other resources failed.");
                    return false;
                }
            }
        }
        if (stepData.getArgoApplicationPath() != null) {
            var getFileRequest = new GetFileRequest(pipelineRepoUrl, stepData.getArgoApplicationPath());
            var argoApplicationConfig = repoManagerApi.getFileFromRepo(getFileRequest, event.getOrgName(), event.getTeamName());
            //TODO: The splitting of the config file should eventually be done on the client side
            for (var applicationConfig : argoApplicationConfig.split("---")) {
                var deployResponse = clientWrapperApi.deploy(event.getOrgName(), ClientWrapperApi.DEPLOY_ARGO_REQUEST, Optional.of(applicationConfig), Optional.empty());
                log.info("Deploying Argo application {}...", stepData.getArgoApplication());
                if (!deployResponse.getSuccess()) {
                    log.error("Deploying the Argo application failed.");
                    return false;
                } else {
                    var watchRequest = new WatchRequest(event.getTeamName(), event.getPipelineName(), stepData.getName(), WATCH_ARGO_APPLICATION_KEY, stepData.getArgoApplication(), deployResponse.getApplicationNamespace());
                    var watching = clientWrapperApi.watchApplication(event.getOrgName(), watchRequest);
                    if (!watching) return false;
                    log.info("Watching Argo application {}", stepData.getArgoApplication());
                }
            }
        } else {
            //TODO: Expectation is argo application has already been created. Not currently supported.
        }
        //TODO: Audit log
        return true;
    }

    private TeamSchema fetchTeamSchema(Event event) {
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
