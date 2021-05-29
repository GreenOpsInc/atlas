package com.greenops.workflowtrigger.api;

import com.greenops.workflowtrigger.api.model.GitCredOpen;
import com.greenops.workflowtrigger.api.model.GitRepoSchema;
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
        assertEquals(pipelineApi.createPipeline(gitRepoSchema), ResponseEntity.ok().build());
    }

    @Test
    public void getPipelineReturnsOk() {
        assertEquals(pipelineApi.getPipeline("getPipelineTest"), ResponseEntity.ok().build());
    }

    @Test
    public void updatePipelineReturnsOk() {
        assertEquals(pipelineApi.updatePipeline("updatePipelineTest", gitRepoSchema), ResponseEntity.ok().build());
    }

    @Test
    public void deletePipelineReturnsOk() {
        assertEquals(pipelineApi.deletePipeline("deletePipelineTest", gitRepoSchema), ResponseEntity.ok().build());
    }
}
