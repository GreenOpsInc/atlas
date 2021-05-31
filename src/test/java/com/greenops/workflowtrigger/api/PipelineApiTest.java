package com.greenops.workflowtrigger.api;

import com.greenops.workflowtrigger.api.model.git.GitCredOpen;
import com.greenops.workflowtrigger.api.model.git.GitRepoSchema;
import com.greenops.workflowtrigger.dbclient.MockDbClient;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.http.ResponseEntity;

import static org.junit.jupiter.api.Assertions.assertEquals;

public class PipelineApiTest {

    private PipelineApi pipelineApi;
    private GitRepoSchema gitRepoSchema;

    @BeforeEach
    void beforeEach() {
        pipelineApi = new PipelineApi(new MockDbClient());
        gitRepoSchema = new GitRepoSchema("https://github.com/argoproj/argocd-example-apps.git", "guestbook/", new GitCredOpen());
    }

    @Test
    public void createPipelineReturnsOk() {
        assertEquals(pipelineApi.createPipeline("team0", "pipeline1", gitRepoSchema), ResponseEntity.ok().build());
    }

    @Test
    public void getPipelineReturnsOk() {
        assertEquals(pipelineApi.getPipeline("team1", "getPipelineTest"), ResponseEntity.ok().build());
    }

    @Test
    public void updatePipelineReturnsOk() {
        assertEquals(pipelineApi.updatePipeline("team2", "updatePipelineTest", gitRepoSchema), ResponseEntity.ok().build());
    }

    @Test
    public void deletePipelineReturnsOk() {
        assertEquals(pipelineApi.deletePipeline("team3", "deletePipelineTest", gitRepoSchema), ResponseEntity.ok().build());
    }
}
