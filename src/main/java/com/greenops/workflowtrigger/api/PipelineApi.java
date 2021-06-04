package com.greenops.workflowtrigger.api;

import com.greenops.workflowtrigger.api.model.git.GitRepoSchema;
import com.greenops.workflowtrigger.api.model.pipeline.TeamSchemaImpl;
import com.greenops.workflowtrigger.api.reposerver.RepoManagerApi;
import com.greenops.workflowtrigger.dbclient.DbClient;
import com.greenops.workflowtrigger.dbclient.DbKey;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@Slf4j
@RestController
@RequestMapping("/")
public class PipelineApi {

    @Autowired
    private DbClient dbClient;

    @Autowired
    private RepoManagerApi repoManagerApi;

    @PostMapping(value = "/team/{orgName}/{parentTeamName}/{teamName}")
    public ResponseEntity<Void> createTeam(@PathVariable("orgName") String orgName,
                                           @PathVariable("parentTeamName") String parentTeamName,
                                           @PathVariable("teamName") String teamName) {
        var key = DbKey.makeDbTeamKey(orgName, teamName);
        if (dbClient.fetchTeamSchema(key) == null) {
            var newTeam = new TeamSchemaImpl(teamName, parentTeamName, orgName);
            if (dbClient.store(key, newTeam)) {
                log.info("Created new team {}", newTeam.getTeamName());
                return ResponseEntity.ok().build();
            } else {
                return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
            }
        }
        return ResponseEntity.badRequest().build();
    }

    @GetMapping(value = "/team/{orgName}/{teamName}")
    public ResponseEntity<Void> readTeam(@PathVariable("orgName") String orgName,
                                         @PathVariable("teamName") String teamName) {
        // TODO: implement team creation logic
        return ResponseEntity.ok().build();
    }

    @PutMapping(value = "/team/{orgName}/{teamName}")
    public ResponseEntity<Void> updateTeam(@PathVariable("orgName") String orgName,
                                           @PathVariable("teamName") String teamName,
                                           @RequestBody UpdateTeamRequest updateTeamRequest) {
        var key = DbKey.makeDbTeamKey(orgName, teamName);
        var teamSchema= dbClient.fetchTeamSchema(key);
        if (teamSchema != null) {
            if (dbClient.store(key, null)) {
                return createTeam(teamSchema.getOrgName(), updateTeamRequest.getNewParentTeamName(), updateTeamRequest.getNewTeamName());
            } else {
                return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
            }
        }
        return createTeam(orgName, updateTeamRequest.getNewParentTeamName(), updateTeamRequest.getNewTeamName());
    }

    @DeleteMapping(value = "/team/{orgName}/{teamName}")
    public ResponseEntity<Void> deleteTeam(@PathVariable("orgName") String orgName,
                                           @PathVariable("teamName") String teamName) {
        // TODO: implement team creation logic
        return ResponseEntity.ok().build();
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

    public PipelineApi(DbClient dbClient) {
        this.dbClient = dbClient;
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
