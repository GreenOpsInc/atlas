package com.greenops.workflowtrigger.api;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.util.datamodel.event.ClientCompletionEvent;
import com.greenops.util.datamodel.event.PipelineTriggerEvent;
import com.greenops.util.datamodel.git.GitCredOpen;
import com.greenops.util.datamodel.git.GitRepoSchema;
import com.greenops.util.datamodel.pipeline.PipelineSchema;
import com.greenops.util.datamodel.pipeline.TeamSchemaImpl;
import com.greenops.util.dbclient.DbClient;
import com.greenops.util.error.AtlasAuthenticationError;
import com.greenops.workflowtrigger.api.reposerver.RepoManagerApi;
import com.greenops.workflowtrigger.dbclient.DbKey;
import com.greenops.workflowtrigger.kafka.KafkaClient;
import com.greenops.workflowtrigger.kubernetesclient.KubernetesClient;
import com.greenops.workflowtrigger.validator.RequestSchemaValidator;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

import static com.greenops.workflowtrigger.api.argoauthenticator.ArgoAuthenticatorApi.*;
import static com.greenops.workflowtrigger.api.reposerver.RepoManagerApiImpl.ROOT_COMMIT;

@Slf4j
@RestController
@RequestMapping("/")
public class PipelineApi {

    private final DbClient dbClient;
    private final KafkaClient kafkaClient;
    private final KubernetesClient kubernetesClient;
    private final RepoManagerApi repoManagerApi;
    private final RequestSchemaValidator requestSchemaValidator;
    private final ObjectMapper objectMapper;

    @Autowired
    public PipelineApi(DbClient dbClient, KafkaClient kafkaClient, KubernetesClient kubernetesClient, RepoManagerApi repoManagerApi, RequestSchemaValidator requestSchemaValidator, ObjectMapper objectMapper) {
        this.dbClient = dbClient;
        this.kafkaClient = kafkaClient;
        this.kubernetesClient = kubernetesClient;
        this.repoManagerApi = repoManagerApi;
        this.requestSchemaValidator = requestSchemaValidator;
        this.objectMapper = objectMapper;
    }

    @PostMapping(value = "/team/{orgName}/{parentTeamName}/{teamName}")
    public ResponseEntity<Void> createTeam(@PathVariable("orgName") String orgName,
                                           @PathVariable("parentTeamName") String parentTeamName,
                                           @PathVariable("teamName") String teamName,
                                           @RequestBody(required = false) List<PipelineSchema> pipelineSchemas) {
        if (!requestSchemaValidator.checkAuthentication()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        }
        var key = DbKey.makeDbTeamKey(orgName, teamName);
        if (dbClient.fetchTeamSchema(key) == null) {
            var newTeam = new TeamSchemaImpl(teamName, parentTeamName, orgName);
            if (pipelineSchemas != null) {
                for (var pipelineSchema : pipelineSchemas) {
                    if (!kubernetesClient.storeGitCred(pipelineSchema.getGitRepoSchema().getGitCred(), DbKey.makeSecretName(orgName, teamName, pipelineSchema.getPipelineName()))) {
                        return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
                    }
                    pipelineSchema.getGitRepoSchema().getGitCred().hide();
                }
                pipelineSchemas.forEach(newTeam::addPipeline);
            }
            dbClient.storeValue(key, newTeam);
            addTeamToOrgList(newTeam.getOrgName(), newTeam.getTeamName());
            log.info("Created new team {}", newTeam.getTeamName());
            return ResponseEntity.ok().build();
        }
        return ResponseEntity.badRequest().build();
    }

    @GetMapping(value = "/team/{orgName}/{teamName}")
    public ResponseEntity<String> readTeam(@PathVariable("orgName") String orgName,
                                           @PathVariable("teamName") String teamName) {
        if (!requestSchemaValidator.checkAuthentication()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        }
        var key = DbKey.makeDbTeamKey(orgName, teamName);
        var teamSchema = dbClient.fetchTeamSchema(key);
        if (teamSchema == null) {
            return ResponseEntity.badRequest().build();
        }
        return ResponseEntity.ok()
                .contentType(MediaType.APPLICATION_JSON)
                .body(schemaToResponsePayload(teamSchema));
    }

    @PutMapping(value = "/team/{orgName}/{teamName}")
    public ResponseEntity<Void> updateTeam(@PathVariable("orgName") String orgName,
                                           @PathVariable("teamName") String teamName,
                                           @RequestBody UpdateTeamRequest updateTeamRequest) {
        if (!requestSchemaValidator.checkAuthentication()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        }
        var key = DbKey.makeDbTeamKey(orgName, teamName);
        var teamSchema = dbClient.fetchTeamSchema(key);
        var response = deleteTeam(orgName, teamName);
        if (!response.getStatusCode().is2xxSuccessful()) {
            return response;
        }
        return createTeam(
                teamSchema.getOrgName(),
                updateTeamRequest.getNewParentTeamName(),
                updateTeamRequest.getNewTeamName(),
                teamSchema.getPipelineSchemas()
        );
    }

    @DeleteMapping(value = "/team/{orgName}/{teamName}")
    public ResponseEntity<Void> deleteTeam(@PathVariable("orgName") String orgName,
                                           @PathVariable("teamName") String teamName) {
        if (!requestSchemaValidator.checkAuthentication()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        }
        var key = DbKey.makeDbTeamKey(orgName, teamName);
        var teamSchema = dbClient.fetchTeamSchema(key);
        for (var pipelineSchema : teamSchema.getPipelineSchemas()) {
            if (!kubernetesClient.storeGitCred(null, DbKey.makeSecretName(orgName, teamName, pipelineSchema.getPipelineName()))) {
                return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
            }
        }
        dbClient.storeValue(key, null);
        removeTeamFromOrgList(orgName, teamName);
        return ResponseEntity.ok().build();
    }

    @PostMapping(value = "/pipeline/{orgName}/{teamName}/{pipelineName}")
    public ResponseEntity<Void> createPipeline(@PathVariable("orgName") String orgName,
                                               @PathVariable("teamName") String teamName,
                                               @PathVariable("pipelineName") String pipelineName,
                                               @RequestBody GitRepoSchema gitRepo) {
        if (!requestSchemaValidator.checkAuthentication()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        }
        var key = DbKey.makeDbTeamKey(orgName, teamName);
        var teamSchema = dbClient.fetchTeamSchema(key);

        if (teamSchema == null) {
            return ResponseEntity.badRequest().build();
        }
        if (teamSchema.getPipelineSchema(pipelineName) != null) {
            return ResponseEntity.status(HttpStatus.CONFLICT).build();
        }

        if (!kubernetesClient.storeGitCred(gitRepo.getGitCred(), DbKey.makeSecretName(orgName, teamName, pipelineName))) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }

        if (!repoManagerApi.cloneRepo(orgName, gitRepo)) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }

        if (!requestSchemaValidator.validateSchemaAccess(orgName, teamName, gitRepo.getGitRepo(), ROOT_COMMIT, CREATE_ACTION, APPLICATION_RESOURCE)) {
            repoManagerApi.deleteRepo(gitRepo);
            return ResponseEntity.status(HttpStatus.FORBIDDEN).build();
        }

        gitRepo.getGitCred().hide();

        teamSchema.addPipeline(pipelineName, gitRepo);

        dbClient.storeValue(key, teamSchema);

        var triggerEvent = new PipelineTriggerEvent(orgName, teamName, pipelineName);
        try {
            generateEvent(objectMapper.writeValueAsString(triggerEvent));
        } catch (JsonProcessingException e) {
            log.error("Serializing the event failed.", e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
        return ResponseEntity.ok().build();
    }

    @GetMapping(value = "/pipeline/{orgName}/{teamName}/{pipelineName}")
    public ResponseEntity<String> getPipeline(@PathVariable("orgName") String orgName,
                                              @PathVariable("teamName") String teamName,
                                              @PathVariable("pipelineName") String pipelineName) {
        if (!requestSchemaValidator.checkAuthentication()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        }
        var key = DbKey.makeDbTeamKey(orgName, teamName);
        var teamSchema = dbClient.fetchTeamSchema(key);
        if (teamSchema == null) {
            return ResponseEntity.badRequest().build();
        }

        var pipelineSchema = teamSchema.getPipelineSchema(pipelineName);
        if (pipelineSchema == null) {
            return ResponseEntity.badRequest().build();
        }

        if (!requestSchemaValidator.validateSchemaAccess(orgName, teamName, pipelineSchema.getGitRepoSchema().getGitRepo(), ROOT_COMMIT, GET_ACTION, APPLICATION_RESOURCE)) {
            return ResponseEntity.status(HttpStatus.FORBIDDEN).build();
        }

        var currentRepoSchema = pipelineSchema.getGitRepoSchema();
        var repoSchema = new GitRepoSchema(currentRepoSchema.getGitRepo(), currentRepoSchema.getPathToRoot(), new GitCredOpen());
        pipelineSchema.setGitRepoSchema(repoSchema);

        return ResponseEntity.ok()
                .contentType(MediaType.APPLICATION_JSON)
                .body(schemaToResponsePayload(pipelineSchema));
    }

    @PutMapping(value = "/pipeline/{orgName}/{teamName}/{pipelineName}")
    public ResponseEntity<Void> updatePipeline(@PathVariable("orgName") String orgName,
                                               @PathVariable("teamName") String teamName,
                                               @PathVariable("pipelineName") String pipelineName,
                                               @RequestBody(required = false) GitRepoSchema gitRepo) {
        if (!requestSchemaValidator.checkAuthentication()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        }
        var response = deletePipeline(orgName, teamName, pipelineName);
        if (!response.getStatusCode().is2xxSuccessful()) {
            return response;
        }
        return createPipeline(
                orgName,
                teamName,
                pipelineName,
                gitRepo
        );
    }

    @DeleteMapping(value = "/pipeline/{orgName}/{teamName}/{pipelineName}")
    public ResponseEntity<Void> deletePipeline(@PathVariable("orgName") String orgName,
                                               @PathVariable("teamName") String teamName,
                                               @PathVariable("pipelineName") String pipelineName) {
        if (!requestSchemaValidator.checkAuthentication()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        }
        var key = DbKey.makeDbTeamKey(orgName, teamName);
        var teamSchema = dbClient.fetchTeamSchema(key);
        if (teamSchema == null) {
            return ResponseEntity.badRequest().build();
        }

        if (teamSchema.getPipelineSchema(pipelineName) == null) {
            return ResponseEntity.badRequest().build();
        }

        if (!requestSchemaValidator.validateSchemaAccess(orgName, teamName, teamSchema.getPipelineSchema(pipelineName).getGitRepoSchema().getGitRepo(), ROOT_COMMIT, DELETE_ACTION, APPLICATION_RESOURCE)) {
            return ResponseEntity.status(HttpStatus.FORBIDDEN).build();
        }

        kubernetesClient.storeGitCred(null, DbKey.makeSecretName(orgName, teamName, pipelineName));

        var gitRepo = teamSchema.getPipelineSchema(pipelineName).getGitRepoSchema();

        teamSchema.removePipeline(pipelineName);

        if (!repoManagerApi.deleteRepo(gitRepo)) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }

        dbClient.storeValue(key, teamSchema);
        return ResponseEntity.ok().build();
    }

    @PostMapping(value = "/sync/{orgName}/{teamName}/{pipelineName}")
    public ResponseEntity<Void> syncPipeline(@PathVariable("orgName") String orgName,
                                             @PathVariable("teamName") String teamName,
                                             @PathVariable("pipelineName") String pipelineName,
                                             @RequestBody GitRepoSchema gitRepo) {
        if (!requestSchemaValidator.checkAuthentication()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        }
        if (!repoManagerApi.sync(gitRepo)) {
            return ResponseEntity.status(HttpStatus.BAD_REQUEST).build();
        }

        if (!requestSchemaValidator.validateSchemaAccess(
                orgName, teamName, gitRepo.getGitRepo(), ROOT_COMMIT,
                SYNC_ACTION, APPLICATION_RESOURCE,
                ACTION_ACTION, CLUSTER_RESOURCE
        )) {
            return ResponseEntity.status(HttpStatus.FORBIDDEN).build();
        }

        var triggerEvent = new PipelineTriggerEvent(orgName, teamName, pipelineName);
        try {
            generateEvent(objectMapper.writeValueAsString(triggerEvent));
        } catch (JsonProcessingException e) {
            log.error("Serializing the event failed.", e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
        return ResponseEntity.ok().build();
    }

    @PostMapping(value = "/client/generateEvent")
    public ResponseEntity<Void> generateEvent(@RequestBody String event) {
        if (!requestSchemaValidator.checkAuthentication()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        }
        kafkaClient.sendMessage(event);
        return ResponseEntity.ok().build();
    }

    private void removeTeamFromOrgList(String orgName, String teamName) {
        var key = DbKey.makeDbListOfTeamsKey(orgName);
        var listOfTeams = dbClient.fetchStringList(key);
        if (listOfTeams == null) listOfTeams = new ArrayList<>();
        listOfTeams = listOfTeams.stream().filter(name -> !name.equals(teamName)).collect(Collectors.toList());
        dbClient.storeValue(key, listOfTeams);
    }

    private void addTeamToOrgList(String orgName, String teamName) {
        var key = DbKey.makeDbListOfTeamsKey(orgName);
        var listOfTeams = dbClient.fetchStringList(key);
        if (listOfTeams == null) listOfTeams = new ArrayList<>();
        if (listOfTeams.stream().noneMatch(name -> name.equals(teamName))) {
            listOfTeams.add(teamName);
            dbClient.storeValue(key, listOfTeams);
        }
    }

    private String schemaToResponsePayload(Object schema) {
        try {
            return objectMapper.writeValueAsString(schema);
        } catch (JsonProcessingException e) {
            throw new RuntimeException("Could not convert schema into response payload.", e);
        }
    }

    private static class UpdateTeamRequest {
        @JsonProperty("teamName")
        private final String teamName;
        @JsonProperty("parentTeamName")
        private final String parentTeamName;

        @JsonCreator
        UpdateTeamRequest(@JsonProperty("teamName") String teamName, @JsonProperty("parentTeamName") String parentTeamName) {
            this.teamName = teamName;
            this.parentTeamName = parentTeamName;
        }

        String getNewTeamName() {
            return teamName;
        }

        String getNewParentTeamName() {
            return parentTeamName;
        }
    }
}
