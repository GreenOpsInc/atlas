package com.greenops.verificationtool.api;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.exc.MismatchedInputException;
import com.greenops.util.datamodel.git.GitRepoSchemaInfo;
import com.greenops.util.datamodel.request.GetFileRequest;
import com.greenops.util.datamodel.pipelinedata.PipelineData;
import com.greenops.verificationtool.datamodel.requests.VerifyPipelineRequestBody;
import com.greenops.verificationtool.datamodel.rules.RuleData;
import com.greenops.verificationtool.datamodel.verification.DAG;
import com.greenops.verificationtool.ingest.apiclient.reposerver.RepoManagerApi;
import com.greenops.verificationtool.ingest.apiclient.workflowtrigger.WorkflowTriggerApi;
import com.greenops.verificationtool.ingest.handling.DagRegistry;
import com.greenops.verificationtool.ingest.handling.RuleEngine;
import com.greenops.verificationtool.ingest.handling.VerificationStatusRegistry;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.io.*;
import java.net.URL;
import java.nio.charset.StandardCharsets;
import java.util.List;
import java.util.UUID;

import static com.greenops.verificationtool.ingest.apiclient.workflowtrigger.WorkflowTriggerApiImpl.NIL;

@RestController
public class VerificationApi {
    public static String ROOT_COMMIT = "ROOT_COMMIT";
    public static String NOT_FOUND = "NOT_FOUND";
    public static String fileName = "pipeline.yaml";
    public static String rulesFilename = "rules.json";
    private final ObjectMapper yamlObjectMapper;
    private final ObjectMapper objectMapper;
    private final DagRegistry dagRegistry;
    private final VerificationStatusRegistry verificationStatusRegistry;
    private final RuleEngine ruleEngine;
    private final RepoManagerApi repoManagerApi;
    private final WorkflowTriggerApi workflowTriggerApi;

    @Autowired
    public VerificationApi(DagRegistry dagRegistry,
                           VerificationStatusRegistry verificationStatusRegistry,
                           RuleEngine ruleEngine,
                           RepoManagerApi repoManagerApi,
                           WorkflowTriggerApi workflowTriggerApi,
                           @Qualifier("objectMapper") ObjectMapper objectMapper,
                           @Qualifier("yamlObjectMapper") ObjectMapper yamlObjectMapper) {
        this.dagRegistry = dagRegistry;
        this.verificationStatusRegistry = verificationStatusRegistry;
        this.ruleEngine = ruleEngine;
        this.objectMapper = objectMapper;
        this.yamlObjectMapper = yamlObjectMapper;
        this.repoManagerApi = repoManagerApi;
        this.workflowTriggerApi = workflowTriggerApi;
    }

    @PostMapping(value = "/verify/{pipelineName}/{orgName}/{parentTeamName}")
    public ResponseEntity<String> verifyPipeline(@PathVariable("pipelineName") String pipelineName,
                                                 @PathVariable("orgName") String orgName,
                                                 @PathVariable("parentTeamName") String parentTeamName,
                                                 @RequestBody String pipelineBody) {
        if (pipelineName == null || pipelineBody == null) {
            return ResponseEntity.badRequest().build();
        }
        try {
            var verifyPipelineRequestBody = this.objectMapper.readValue(pipelineBody, VerifyPipelineRequestBody.class);
            var teamName = verifyPipelineRequestBody.getTeamName();
            if (teamName == null) {
                teamName = String.valueOf(UUID.randomUUID());
            }
            var gitRepoUrl = verifyPipelineRequestBody.getGitRepoUrl();
            var pathToRoot = verifyPipelineRequestBody.getPathToRoot();
            var pipelineIdentifier = pipelineName + "-" + teamName;

            var rules = getRules(gitRepoUrl, pathToRoot);
            if (rules != null) {
                this.ruleEngine.registerRules(pipelineIdentifier, rules);
            }

            var gitRepoSchemaInfo = new GitRepoSchemaInfo(gitRepoUrl, pathToRoot);
            var getFileRequest = new GetFileRequest(gitRepoSchemaInfo, fileName, ROOT_COMMIT);
            var yamlPipeline = this.repoManagerApi.getFileFromRepo(getFileRequest, orgName, teamName);
            var pipelineObj = objectMapper.readValue(
                    objectMapper.writeValueAsString(
                            yamlObjectMapper.readValue(yamlPipeline, Object.class)
                    ),
                    PipelineData.class);

            DAG dag = new DAG(pipelineObj, pipelineName);
            this.verificationStatusRegistry.putVerificationStatus(pipelineIdentifier);
            this.dagRegistry.registerDAG(pipelineIdentifier, dag);

            if (this.workflowTriggerApi.readTeam(orgName, teamName).equals(NIL)) {
                this.workflowTriggerApi.createTeam(orgName, parentTeamName, teamName);
            }
            if (this.workflowTriggerApi.getPipelineEndpoint(orgName, teamName, pipelineName).equals(NIL)) {
                this.workflowTriggerApi.createPipeline(orgName, pipelineName, teamName, gitRepoUrl, pathToRoot);
            }
            this.workflowTriggerApi.syncPipeline(orgName, pipelineName, teamName, gitRepoUrl, pathToRoot);
            return ResponseEntity.ok().body(schemaToResponsePayload(pipelineObj));
        } catch (IOException e) {
            return ResponseEntity.status(400).body(e.getMessage());
        }
    }

    @GetMapping(value = "/status/{pipelineName}/{teamName}")
    public ResponseEntity<String> getSinglePipelineStatus(@PathVariable("pipelineName") String pipelineName, @PathVariable("teamName") String teamName) {
        if (pipelineName == null) {
            return ResponseEntity.badRequest().build();
        }
        var pipelineStatus = this.verificationStatusRegistry.getVerificationStatus(pipelineName + "-" + teamName);
        if (pipelineStatus == null) {
            return ResponseEntity.ok().contentType(MediaType.APPLICATION_JSON).body("Pipeline not found");
        }
        return ResponseEntity.ok().contentType(MediaType.APPLICATION_JSON).body(schemaToResponsePayload(pipelineStatus));
    }

    @GetMapping(value = "/status/all")
    public ResponseEntity<String> getPipelineStatus() {
        return ResponseEntity.ok().contentType(MediaType.APPLICATION_JSON).body(schemaToResponsePayload(this.verificationStatusRegistry));
    }

    private List<RuleData> getRules(String gitRepoUrl, String pathToRoot) throws IOException {
        var contentUrl = gitRepoUrl.replace(".git", "").replace("github.com", "raw.githubusercontent.com");
        var url = String.format("%s/main/%s%s", contentUrl, pathToRoot, rulesFilename);
        InputStream inputStream = new URL(url).openStream();
        try {
            BufferedReader rd = new BufferedReader(new InputStreamReader(inputStream, StandardCharsets.UTF_8));
            String jsonText = readAll(rd);
            return objectMapper.readValue(jsonText, new TypeReference<List<RuleData>>() {
            });
        } catch (IOException e) {
            System.out.println(e.getMessage());
        } finally {
            inputStream.close();
        }
        return null;
    }

    private static String readAll(Reader rd) throws IOException {
        StringBuilder sb = new StringBuilder();
        int cp;
        while ((cp = rd.read()) != -1) {
            sb.append((char) cp);
        }
        return sb.toString();
    }

    private String schemaToResponsePayload(Object schema) {
        try {
            return objectMapper.writeValueAsString(schema);
        } catch (JsonProcessingException e) {
            throw new RuntimeException("Could not convert schema into response payload.", e);
        }
    }
}