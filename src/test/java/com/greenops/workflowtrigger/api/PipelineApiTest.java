package com.greenops.workflowtrigger.api;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.workflowtrigger.api.model.git.GitCredMachineUser;
import com.greenops.workflowtrigger.api.model.git.GitCredOpen;
import com.greenops.workflowtrigger.api.model.git.GitCredToken;
import com.greenops.workflowtrigger.api.model.git.GitRepoSchema;
import com.greenops.workflowtrigger.api.model.mixin.git.GitCredMachineUserMixin;
import com.greenops.workflowtrigger.api.model.mixin.git.GitCredTokenMixin;
import com.greenops.workflowtrigger.api.model.mixin.git.GitRepoSchemaMixin;
import com.greenops.workflowtrigger.api.model.mixin.pipeline.PipelineSchemaMixin;
import com.greenops.workflowtrigger.api.model.mixin.pipeline.TeamSchemaMixin;
import com.greenops.workflowtrigger.api.model.pipeline.PipelineSchemaImpl;
import com.greenops.workflowtrigger.api.model.pipeline.TeamSchemaImpl;
import com.greenops.workflowtrigger.api.reposerver.RepoManagerApi;
import com.greenops.workflowtrigger.dbclient.DbClient;
import com.greenops.workflowtrigger.dbclient.DbKey;
import com.greenops.workflowtrigger.kafka.KafkaClient;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.mockito.ArgumentMatchers.any;

public class PipelineApiTest {

    private PipelineApi pipelineApi;
    private GitRepoSchema gitRepoSchema;
    private ObjectMapper objectMapper;

    // TODO: probably need a cleaner/better way to test the result of getPipeline than using the object mapper instance
    //  / and its expected output JSON as variables here
    private String pipelineSchemaJson;

    @BeforeEach
    void beforeEach() throws JsonProcessingException {
        var teamSchemaNew = new TeamSchemaImpl("team1", "root", "org name");
        var teamSchemaOld = new TeamSchemaImpl("team2", "root", "org name");

        gitRepoSchema = new GitRepoSchema("https://github.com/argoproj/argocd-example-apps.git", "guestbook/", new GitCredOpen());


        var pipelineSchema = new PipelineSchemaImpl("pipeline1", gitRepoSchema);

        teamSchemaOld.addPipeline(pipelineSchema);


        var dbClient = Mockito.mock(DbClient.class);
        Mockito.when(dbClient.fetchTeamSchema(DbKey.makeDbTeamKey("org name", "team1"))).thenReturn(teamSchemaNew);
        Mockito.when(dbClient.fetchTeamSchema(DbKey.makeDbTeamKey("org name", "team2"))).thenReturn(teamSchemaOld);
        Mockito.when(dbClient.store(Mockito.anyString(), any())).thenReturn(true);

        var kafkaClient = Mockito.mock(KafkaClient.class);
        Mockito.doNothing().when(kafkaClient).sendMessage(any(String.class));


        var repoManagerApi = Mockito.mock(RepoManagerApi.class);
        Mockito.when(repoManagerApi.cloneRepo(any())).thenReturn(true);
        Mockito.when(repoManagerApi.deleteRepo(any())).thenReturn(true);

        objectMapper = new ObjectMapper()
                .addMixIn(TeamSchemaImpl.class, TeamSchemaMixin.class)
                .addMixIn(PipelineSchemaImpl.class, PipelineSchemaMixin.class)
                .addMixIn(GitRepoSchema.class, GitRepoSchemaMixin.class)
                .addMixIn(GitCredMachineUser.class, GitCredMachineUserMixin.class)
                .addMixIn(GitCredToken.class, GitCredTokenMixin.class);

        pipelineSchemaJson = objectMapper.writeValueAsString(pipelineSchema);

        pipelineApi = new PipelineApi(dbClient, kafkaClient, repoManagerApi, objectMapper);

    }

    @Test
    public void createPipelineReturnsOk() {
        assertEquals(pipelineApi.createPipeline("org name", "team1", "pipeline1", gitRepoSchema), ResponseEntity.ok().build());
    }

    @Test
    public void getPipelineReturnsOk() {
        assertEquals(pipelineApi.getPipeline("org name", "team2", "pipeline1"), ResponseEntity.ok()
                .contentType(MediaType.APPLICATION_JSON)
                .body(pipelineSchemaJson));
    }

    @Test
    public void updatePipelineReturnsOk() {
        assertEquals(pipelineApi.updatePipeline("org name", "team2", "pipeline1", gitRepoSchema), ResponseEntity.ok().build());
    }

    @Test
    public void deletePipelineReturnsOk() {
        assertEquals(pipelineApi.deletePipeline("org name", "team2", "pipeline1", gitRepoSchema), ResponseEntity.ok().build());
    }
}
