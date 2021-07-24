package com.greenops.workflowtrigger.api;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.workflowtrigger.api.model.event.ClientCompletionEvent;
import com.greenops.workflowtrigger.api.model.git.GitCredOpen;
import com.greenops.workflowtrigger.api.model.git.GitRepoSchema;
import com.greenops.workflowtrigger.api.model.pipeline.PipelineSchema;
import com.greenops.workflowtrigger.api.model.pipeline.TeamSchemaImpl;
import com.greenops.workflowtrigger.api.reposerver.RepoManagerApi;
import com.greenops.workflowtrigger.dbclient.DbClient;
import com.greenops.workflowtrigger.dbclient.DbKey;
import com.greenops.workflowtrigger.kafka.KafkaClient;
import com.greenops.workflowtrigger.kubernetesclient.KubernetesClient;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

@Slf4j
@RestController
@RequestMapping("/")
public class PipelineApi {

    private final DbClient dbClient;
    private final KafkaClient kafkaClient;
    private final KubernetesClient kubernetesClient;
    private final RepoManagerApi repoManagerApi;
    private final ObjectMapper objectMapper;

    @Autowired
    public PipelineApi(DbClient dbClient, KafkaClient kafkaClient, KubernetesClient kubernetesClient, RepoManagerApi repoManagerApi, ObjectMapper objectMapper) {
        this.dbClient = dbClient;
        this.kafkaClient = kafkaClient;
        this.kubernetesClient = kubernetesClient;
        this.repoManagerApi = repoManagerApi;
        this.objectMapper = objectMapper;
    }

    @PostMapping(value = "/team/{orgName}/{parentTeamName}/{teamName}")
    public ResponseEntity<Void> createTeam(@PathVariable("orgName") String orgName,
                                           @PathVariable("parentTeamName") String parentTeamName,
                                           @PathVariable("teamName") String teamName,
                                           @RequestBody(required = false) List<PipelineSchema> pipelineSchemas) {
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
            if (dbClient.store(key, newTeam)) {
                addTeamToOrgList(newTeam.getOrgName(), newTeam.getTeamName());
                log.info("Created new team {}", newTeam.getTeamName());
                return ResponseEntity.ok().build();
            } else {
                return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
            }
        }
        return ResponseEntity.badRequest().build();
    }

    @GetMapping(value = "/team/{orgName}/{teamName}")
    public ResponseEntity<String> readTeam(@PathVariable("orgName") String orgName,
                                           @PathVariable("teamName") String teamName) {
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
        var key = DbKey.makeDbTeamKey(orgName, teamName);
        var teamSchema = dbClient.fetchTeamSchema(key);
        for (var pipelineSchema : teamSchema.getPipelineSchemas()) {
            if (!kubernetesClient.storeGitCred(null, DbKey.makeSecretName(orgName, teamName, pipelineSchema.getPipelineName()))) {
                return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
            }
        }
        if (dbClient.store(key, null)) {
            removeTeamFromOrgList(orgName, teamName);
            return ResponseEntity.ok().build();
        } else {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }

    @PostMapping(value = "/pipeline/{orgName}/{teamName}/{pipelineName}")
    public ResponseEntity<Void> createPipeline(@PathVariable("orgName") String orgName,
                                               @PathVariable("teamName") String teamName,
                                               @PathVariable("pipelineName") String pipelineName,
                                               @RequestBody GitRepoSchema gitRepo) {
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

        if (!repoManagerApi.cloneRepo(gitRepo)) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }

        gitRepo.getGitCred().hide();

        teamSchema.addPipeline(pipelineName, gitRepo);

        if (!dbClient.store(key, teamSchema)) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }

        //TODO: TriggerEvents don't need all of this information. This should be replaced with a special type called a "TriggerEvent"
        var triggerEvent = new ClientCompletionEvent("Healthy", orgName, teamName, pipelineName, "ATLAS_ROOT_DATA", "", "", "", gitRepo.getGitRepo());
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
        var key = DbKey.makeDbTeamKey(orgName, teamName);
        var teamSchema = dbClient.fetchTeamSchema(key);
        if (teamSchema == null) {
            return ResponseEntity.badRequest().build();
        }

        var pipelineSchema = teamSchema.getPipelineSchema(pipelineName);
        if (pipelineSchema == null) {
            return ResponseEntity.badRequest().build();
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
        var response = deletePipeline(orgName, teamName, pipelineName, gitRepo);
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
                                               @PathVariable("pipelineName") String pipelineName,
                                               @RequestBody GitRepoSchema gitRepo) {
        var key = DbKey.makeDbTeamKey(orgName, teamName);
        var teamSchema = dbClient.fetchTeamSchema(key);
        if (teamSchema == null) {
            return ResponseEntity.badRequest().build();
        }

        if (teamSchema.getPipelineSchema(pipelineName) == null) {
            return ResponseEntity.badRequest().build();
        }

        kubernetesClient.storeGitCred(null, DbKey.makeSecretName(orgName, teamName, pipelineName));

        teamSchema.removePipeline(pipelineName);

        if (!repoManagerApi.deleteRepo(gitRepo)) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }

        if (dbClient.store(key, teamSchema)) {
            return ResponseEntity.ok().build();
        }
        return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
    }

    @PostMapping(value = "/sync/{orgName}/{teamName}/{pipelineName}")
    public ResponseEntity<Void> syncPipeline(@PathVariable("orgName") String orgName,
                                             @PathVariable("teamName") String teamName,
                                             @PathVariable("pipelineName") String pipelineName,
                                             @RequestBody GitRepoSchema gitRepo) {
        if (!repoManagerApi.sync(gitRepo)) {
            return ResponseEntity.status(HttpStatus.BAD_REQUEST).build();
        }

        //TODO: TriggerEvents don't need all of this information. This should be replaced with a special type called a "TriggerEvent"
        var triggerEvent = new ClientCompletionEvent("Healthy", orgName, teamName, pipelineName, "ATLAS_ROOT_DATA", "", "", "", gitRepo.getGitRepo());
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
        kafkaClient.sendMessage(event);
        return ResponseEntity.ok().build();
    }

    private void removeTeamFromOrgList(String orgName, String teamName) {
        var key = DbKey.makeDbListOfTeamsKey(orgName);
        var status = false;
        while (!status) {
            var listOfTeams = dbClient.fetchList(key);
            if (listOfTeams == null) listOfTeams = new ArrayList<>();
            listOfTeams = listOfTeams.stream().filter(name -> !name.equals(teamName)).collect(Collectors.toList());
            status = dbClient.store(key, listOfTeams);
        }
    }

    private void addTeamToOrgList(String orgName, String teamName) {
        var key = DbKey.makeDbListOfTeamsKey(orgName);
        var status = false;
        while (!status) {
            var listOfTeams = dbClient.fetchList(key);
            if (listOfTeams == null) listOfTeams = new ArrayList<>();
            if (listOfTeams.stream().noneMatch(name -> name.equals(teamName))) {
                listOfTeams.add(teamName);
                status = dbClient.store(key, listOfTeams);

            } else {
                status = true;
            }
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
        private final String teamName;
        private final String parentTeamName;

        UpdateTeamRequest(String teamName, String parentTeamName) {
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
