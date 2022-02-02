package com.greenops.verificationtool.api;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.util.datamodel.git.GitRepoSchemaInfo;
import com.greenops.util.datamodel.request.GetFileRequest;
import com.greenops.util.datamodel.pipelinedata.PipelineData;
import com.greenops.verificationtool.datamodel.requests.VerifyPipelineRequestBody;
import com.greenops.verificationtool.datamodel.verification.DAG;
import com.greenops.verificationtool.ingest.apiclient.reposerver.RepoManagerApi;
import com.greenops.verificationtool.ingest.apiclient.workflowtrigger.WorkflowTriggerApi;
import com.greenops.verificationtool.ingest.handling.DagRegistry;
import com.greenops.verificationtool.ingest.handling.RuleEngine;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;


@RestController
public class VerificationApi {
    public static String ROOT_COMMIT = "ROOT_COMMIT";
    public static String NOT_FOUND = "NOT_FOUND";
    public static String fileName = "pipeline.yaml";
    private final ObjectMapper yamlObjectMapper;
    private final ObjectMapper objectMapper;
    private final DagRegistry dagRegistry;
    private final RuleEngine ruleEngine;
    private final RepoManagerApi repoManagerApi;
    private final WorkflowTriggerApi workflowTriggerApi;

    @Autowired
    public VerificationApi(DagRegistry dagRegistry,
                           RuleEngine ruleEngine,
                           RepoManagerApi repoManagerApi,
                           WorkflowTriggerApi workflowTriggerApi,
                           @Qualifier("objectMapper") ObjectMapper objectMapper,
                           @Qualifier("yamlObjectMapper") ObjectMapper yamlObjectMapper) {
        this.dagRegistry = dagRegistry;
        this.ruleEngine = ruleEngine;
        this.objectMapper = objectMapper;
        this.yamlObjectMapper = yamlObjectMapper;
        this.repoManagerApi = repoManagerApi;
        this.workflowTriggerApi = workflowTriggerApi;
    }

    @PostMapping(value = "/verify/{pipelineName}/{teamName}/{orgName}/{parentTeamName}")
    public ResponseEntity<String> verifyPipeline(@PathVariable("pipelineName") String pipelineName,
                                                 @PathVariable("teamName") String teamName,
                                                 @PathVariable("orgName") String orgName,
                                                 @PathVariable("parentTeamName") String parentTeamName,
                                                 @RequestBody String pipelineBody) {
        if (pipelineName == null || teamName == null || pipelineBody == null) {
            return ResponseEntity.badRequest().build();
        }
        try {
            var verifyPipelineRequestBody = this.objectMapper.readValue(pipelineBody, VerifyPipelineRequestBody.class);
            var gitRepoUrl = verifyPipelineRequestBody.getGitRepoUrl();
            var pathToRoot = verifyPipelineRequestBody.getPathToRoot();
            this.ruleEngine.registerRules(pipelineName, verifyPipelineRequestBody.getRules());

            this.workflowTriggerApi.createTeam(orgName, parentTeamName, teamName);

            var gitRepoSchemaInfo = new GitRepoSchemaInfo(gitRepoUrl, pathToRoot);
            var getFileRequest = new GetFileRequest(gitRepoSchemaInfo, fileName, ROOT_COMMIT);
            var yamlPipeline = this.repoManagerApi.getFileFromRepo(getFileRequest, orgName, teamName);
            var pipelineObj = objectMapper.readValue(
                    objectMapper.writeValueAsString(
                            yamlObjectMapper.readValue(yamlPipeline, Object.class)
                    ),
                    PipelineData.class);

            DAG dag = new DAG(pipelineObj, pipelineName);
            this.dagRegistry.registerDAG(pipelineName, dag);

            this.workflowTriggerApi.createPipeline(orgName, pipelineName, teamName, gitRepoUrl);
            this.workflowTriggerApi.syncPipeline(orgName, pipelineName, teamName, gitRepoUrl);
            return ResponseEntity.ok().body(schemaToResponsePayload(pipelineObj));
        } catch (JsonProcessingException e) {
            return ResponseEntity.status(400).body(e.getMessage());
        }
    }

    @GetMapping(value = "/status/{pipelineName}")
    public ResponseEntity<String> getSinglePipelineStatus(@PathVariable("pipelineName") String pipelineName) {
        if (pipelineName == null) {
            return ResponseEntity.badRequest().build();
        }
        var pipelineStatus = this.dagRegistry.getSinglePipelineStatus(pipelineName);
        if (pipelineStatus.equals(NOT_FOUND)) {
            return ResponseEntity.ok().contentType(MediaType.APPLICATION_JSON).body("Pipeline not found");
        }
        return ResponseEntity.ok().contentType(MediaType.APPLICATION_JSON).body(pipelineStatus);
    }

    @GetMapping(value = "/status/all")
    public ResponseEntity<String> getPipelineStatus() {
        var pipelinesStatus = this.dagRegistry.getPipelineStatus();
        return ResponseEntity.ok().contentType(MediaType.APPLICATION_JSON).body(schemaToResponsePayload(pipelinesStatus));
    }

    private String schemaToResponsePayload(Object schema) {
        try {
            return objectMapper.writeValueAsString(schema);
        } catch (JsonProcessingException e) {
            throw new RuntimeException("Could not convert schema into response payload.", e);
        }
    }
}