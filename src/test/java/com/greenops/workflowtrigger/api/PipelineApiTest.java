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
import com.greenops.workflowtrigger.api.model.pipeline.PipelineSchema;
import com.greenops.workflowtrigger.api.model.pipeline.PipelineSchemaImpl;
import com.greenops.workflowtrigger.api.model.pipeline.TeamSchema;
import com.greenops.workflowtrigger.api.model.pipeline.TeamSchemaImpl;
import com.greenops.workflowtrigger.api.reposerver.RepoManagerApi;
import com.greenops.workflowtrigger.dbclient.DbClient;
import com.greenops.workflowtrigger.dbclient.DbKey;
import com.greenops.workflowtrigger.kafka.KafkaClient;
import com.greenops.workflowtrigger.kubernetesclient.KubernetesClient;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.mockito.ArgumentMatchers.any;

import static org.mockito.Mockito.when;


public class PipelineApiTest {

    private PipelineApi pipelineApi;
    private GitRepoSchema gitRepoSchema;
    private ObjectMapper objectMapper;
    private String pipelineSchemaJson;
    private DbClient dbClient;
    private TeamSchema teamSchemaNew;
    private TeamSchema teamSchemaOld;
    private PipelineSchema pipelineSchema;
    private KafkaClient kafkaClient;
    private RepoManagerApi repoManagerApi;

    @BeforeEach
    void beforeEach() throws JsonProcessingException {
        teamSchemaNew = new TeamSchemaImpl("team1", "root", "org name");
        teamSchemaOld = new TeamSchemaImpl("team2", "root", "org name");
        gitRepoSchema = new GitRepoSchema("https://github.com/argoproj/argocd-example-apps.git", "guestbook/", new GitCredOpen());
        pipelineSchema = new PipelineSchemaImpl("pipeline1", gitRepoSchema);

        teamSchemaOld.addPipeline(pipelineSchema);


        dbClient = Mockito.mock(DbClient.class);

        kafkaClient = Mockito.mock(KafkaClient.class);
        Mockito.doNothing().when(kafkaClient).sendMessage(any(String.class));

        var kubernetesClient = Mockito.mock(KubernetesClient.class);
        Mockito.when(kubernetesClient.storeGitCred(any(), any())).thenReturn(true);
        Mockito.when(kubernetesClient.fetchGitCred(any())).thenReturn(null);

        repoManagerApi = Mockito.mock(RepoManagerApi.class);
        Mockito.when(repoManagerApi.cloneRepo(any())).thenReturn(true);
        Mockito.when(repoManagerApi.deleteRepo(any())).thenReturn(true);

        objectMapper = new ObjectMapper()
                .addMixIn(TeamSchemaImpl.class, TeamSchemaMixin.class)
                .addMixIn(PipelineSchemaImpl.class, PipelineSchemaMixin.class)
                .addMixIn(GitRepoSchema.class, GitRepoSchemaMixin.class)
                .addMixIn(GitCredMachineUser.class, GitCredMachineUserMixin.class)
                .addMixIn(GitCredToken.class, GitCredTokenMixin.class);

        pipelineSchemaJson = objectMapper.writeValueAsString(pipelineSchema);

        pipelineApi = new PipelineApi(dbClient, kafkaClient, kubernetesClient, repoManagerApi, objectMapper);

    }

    @Test
    public void createPipelineReturnsOk() {
        when(dbClient.store(Mockito.anyString(), any())).thenReturn(true);
        when(dbClient.fetchTeamSchema(DbKey.makeDbTeamKey("org name", "team1"))).thenReturn(teamSchemaNew);
        when(repoManagerApi.cloneRepo(any())).thenReturn(true);
        assertEquals(pipelineApi.createPipeline("org name", "team1", "pipeline1", gitRepoSchema), ResponseEntity.ok().build());
    }

    @Test
    public void getPipelineReturnsOk() {
        when(dbClient.fetchTeamSchema(DbKey.makeDbTeamKey("org name", "team2"))).thenReturn(teamSchemaOld);
        assertEquals(pipelineApi.getPipeline("org name", "team2", "pipeline1"), ResponseEntity.ok()
                .contentType(MediaType.APPLICATION_JSON)
                .body(pipelineSchemaJson));
    }

    @Test
    public void updatePipelineReturnsOk() {
        when(dbClient.store(Mockito.anyString(), any())).thenReturn(true);
        when(dbClient.fetchTeamSchema(DbKey.makeDbTeamKey("org name", "team2"))).thenReturn(teamSchemaOld);
        when(repoManagerApi.cloneRepo(any())).thenReturn(true);
        assertEquals(pipelineApi.updatePipeline("org name", "team2", "pipeline1", gitRepoSchema), ResponseEntity.ok().build());
    }

    @Test
    public void deletePipelineReturnsOk() {
        when(dbClient.store(Mockito.anyString(), any())).thenReturn(true);
        when(dbClient.fetchTeamSchema(DbKey.makeDbTeamKey("org name", "team2"))).thenReturn(teamSchemaOld);
        when(repoManagerApi.deleteRepo(any())).thenReturn(true);
        assertEquals(pipelineApi.deletePipeline("org name", "team2", "pipeline1", gitRepoSchema), ResponseEntity.ok().build());
    }

    @Test
    public void createPipelineFailsWhenTeamDoesNotExist() {
        when(dbClient.store(Mockito.anyString(), any())).thenReturn(true);
        when(dbClient.fetchTeamSchema(any())).thenReturn(null);
        assertEquals(pipelineApi.createPipeline("org name", "team3", "pipeline2", gitRepoSchema), ResponseEntity.badRequest().build());

    }

    @Test
    public void createPipelineFailsWhenPipelineExists() {
        when(dbClient.fetchTeamSchema(DbKey.makeDbTeamKey("org name", "team2"))).thenReturn(teamSchemaOld);
        assertEquals(pipelineApi.createPipeline("org name", "team2", "pipeline1", gitRepoSchema), ResponseEntity.status(HttpStatus.CONFLICT).build());
    }

    @Test
    public void createPipelineFailsWhenRepoManagerFails() {
        when(dbClient.fetchTeamSchema(DbKey.makeDbTeamKey("org name", "team1"))).thenReturn(teamSchemaNew);
        when(repoManagerApi.cloneRepo(gitRepoSchema)).thenReturn(false);
        assertEquals(pipelineApi.createPipeline("org name", "team1", "pipeline1", gitRepoSchema), ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build());
    }

    @Test
    public void createPipelineFailsWhenDbClientFails() {
        when(dbClient.fetchTeamSchema(DbKey.makeDbTeamKey("org name", "team1"))).thenReturn(teamSchemaNew);
        when(dbClient.store(any(), any())).thenReturn(false);
        assertEquals(pipelineApi.createPipeline("org name", "team1", "pipeline1", gitRepoSchema), ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build());
    }

    @Test
    public void getPipelineFailsWhenTeamDoesNotExist() {
        when(dbClient.fetchTeamSchema(any())).thenReturn(null);
        assertEquals(pipelineApi.getPipeline("org name", "team3", "pipeline1"), ResponseEntity.badRequest().build());
    }

    @Test
    public void getPipelineFailsWhenPipelineDoesNotExist() {
        when(dbClient.fetchTeamSchema(DbKey.makeDbTeamKey("org name", "team2"))).thenReturn(teamSchemaOld);
        assertEquals(pipelineApi.getPipeline("org name", "team2", "pipeline2"), ResponseEntity.badRequest().build());
    }

    @Test
    public void deletePipelineFailsWhenTeamDoesNotExist() {
        when(dbClient.fetchTeamSchema(any())).thenReturn(null);
        assertEquals(pipelineApi.deletePipeline("org name", "team3", "pipeline1", gitRepoSchema), ResponseEntity.badRequest().build());
    }

    @Test
    public void deletePipelineFailsWhenPipelineDoesNotExist() {
        when(dbClient.fetchTeamSchema(DbKey.makeDbTeamKey("org name", "team2"))).thenReturn(teamSchemaOld);
        assertEquals(pipelineApi.deletePipeline("org name", "team2", "pipeline2", gitRepoSchema), ResponseEntity.badRequest().build());
    }

    @Test
    public void deletePipelineFailsWhenRepoManagerFails() {
        when(dbClient.fetchTeamSchema(DbKey.makeDbTeamKey("org name", "team2"))).thenReturn(teamSchemaOld);
        when(repoManagerApi.deleteRepo(gitRepoSchema)).thenReturn(false);
        assertEquals(pipelineApi.deletePipeline("org name", "team2", "pipeline1", gitRepoSchema), ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build());
    }

    @Test
    public void deletePipelineFailsWhenDbClientFails() {
        when(dbClient.fetchTeamSchema(DbKey.makeDbTeamKey("org name", "team2"))).thenReturn(teamSchemaOld);
        when(dbClient.store(any(), any())).thenReturn(false);
        assertEquals(pipelineApi.deletePipeline("org name", "team2", "pipeline1", gitRepoSchema), ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build());
    }
}
