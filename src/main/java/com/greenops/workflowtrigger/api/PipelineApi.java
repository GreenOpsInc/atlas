package com.greenops.workflowtrigger.api;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.workflowtrigger.api.model.git.GitCredMachineUser;
import com.greenops.workflowtrigger.api.model.git.GitCredOpen;
import com.greenops.workflowtrigger.api.model.git.GitRepoSchema;
import com.greenops.workflowtrigger.api.model.mixin.git.GitCredMachineUserMixin;
import com.greenops.workflowtrigger.api.model.mixin.git.GitRepoSchemaMixin;
import com.greenops.workflowtrigger.api.model.mixin.pipeline.PipelineSchemaMixin;
import com.greenops.workflowtrigger.api.model.mixin.pipeline.TeamSchemaMixin;
import com.greenops.workflowtrigger.api.model.pipeline.PipelineSchema;
import com.greenops.workflowtrigger.api.model.pipeline.PipelineSchemaImpl;
import com.greenops.workflowtrigger.api.model.pipeline.TeamSchemaImpl;
import com.greenops.workflowtrigger.api.reposerver.RepoManagerApi;
import com.greenops.workflowtrigger.dbclient.DbClient;
import com.greenops.workflowtrigger.dbclient.DbKey;
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
    private final RepoManagerApi repoManagerApi;
    private final ObjectMapper objectMapper;

    @Autowired
    public PipelineApi(DbClient dbClient, RepoManagerApi repoManagerApi) {
        this.dbClient = dbClient;
        this.repoManagerApi = repoManagerApi;
        objectMapper = new ObjectMapper()
                .addMixIn(TeamSchemaImpl.class, TeamSchemaMixin.class)
                .addMixIn(PipelineSchemaImpl.class, PipelineSchemaMixin.class)
                .addMixIn(GitRepoSchema.class, GitRepoSchemaMixin.class)
                .addMixIn(GitCredMachineUser.class, GitCredMachineUserMixin.class);
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
        for (var pipelineSchema : teamSchema.getPipelineSchemas()) {
            pipelineSchema.getGitRepoSchema().setGitCred(new GitCredOpen());
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
        var teamSchema= dbClient.fetchTeamSchema(key);
        var response= deleteTeam(orgName, teamName);
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
        if (dbClient.store(key,null)) {
            removeTeamFromOrgList(orgName, teamName);
            return ResponseEntity.ok().build();
        } else {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }

    @PostMapping(value = "/pipeline/{teamName}/{pipelineName}")
    public ResponseEntity<Void> createPipeline(@PathVariable("teamName") String teamName,
                                               @PathVariable("pipelineName") String pipelineName,
                                               @RequestBody GitRepoSchema gitRepo) {
        // TODO: implement pipeline creation logic
        return ResponseEntity.ok().build();
    }

    @GetMapping(value = "/pipeline/{teamName}/{pipelineName}")
    public ResponseEntity<GitRepoSchema> getPipeline(@PathVariable("teamName") String teamName, @PathVariable("pipelineName") String pipelineName) {
        // TODO: implement pipeline fetch logic
        return ResponseEntity.ok().build();
    }

    @PutMapping(value = "/pipeline/{teamName}/{pipelineName}")
    public ResponseEntity<Void> updatePipeline(@PathVariable("teamName") String teamName, @PathVariable("pipelineName") String pipelineName, @RequestBody(required = false) GitRepoSchema gitRepo) {
        // TODO: implement pipeline update logic
        return ResponseEntity.ok().build();
    }

    @DeleteMapping(value = "/pipeline/{teamName}/{pipelineName}")
    public ResponseEntity<Void> deletePipeline(@PathVariable("teamName") String teamName, @PathVariable("pipelineName") String pipelineName, @RequestBody GitRepoSchema gitRepo) {
        // TODO: implement pipeline deletion logic
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
