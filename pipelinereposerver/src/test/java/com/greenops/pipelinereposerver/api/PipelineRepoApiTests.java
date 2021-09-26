package com.greenops.pipelinereposerver.api;

import com.greenops.pipelinereposerver.repomanager.RepoManager;
import com.greenops.util.datamodel.git.GitCredOpen;
import com.greenops.util.datamodel.git.GitRepoSchema;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.mockito.Mockito.when;

public class PipelineRepoApiTests {
    private RepoManager repoManager;
    private PipelineRepoApi pipelineApi;
    private GitRepoSchema schema;

    @BeforeEach
    public void beforeEach() {
        repoManager = Mockito.mock(RepoManager.class);
        schema = new GitRepoSchema("https://test.git", "/", new GitCredOpen());
        pipelineApi = new PipelineRepoApi(repoManager);
    }

    @Test
    public void testCloneWorksWhenGitRepoDoesNotExist() {
        when(repoManager.clone(schema)).thenReturn(true);
        when(repoManager.containsGitRepoSchema(schema)).thenReturn(false);
        var response = pipelineApi.cloneRepo("org", schema);

        assertEquals(ResponseEntity.ok().build(), response);
    }

    @Test
    public void testCloneFailsWhenGitRepoDoesNotExist() {
        when(repoManager.clone(schema)).thenReturn(false);
        when(repoManager.containsGitRepoSchema(schema)).thenReturn(false);
        var response = pipelineApi.cloneRepo("org", schema);

        assertEquals(ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build(), response);
    }

    @Test
    public void testCloneWorksWhenGitRepoExists() {
        when(repoManager.update(schema)).thenReturn(true);
        when(repoManager.containsGitRepoSchema(schema)).thenReturn(true);
        var response = pipelineApi.cloneRepo("org", schema);

        assertEquals(ResponseEntity.ok().build(), response);
    }

    @Test
    public void testCloneFailsWhenGitRepoExists() {
        when(repoManager.update(schema)).thenReturn(false);
        when(repoManager.containsGitRepoSchema(schema)).thenReturn(true);
        var response = pipelineApi.cloneRepo("org", schema);

        assertEquals(ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build(), response);
    }

    @Test
    public void testDeleteRepoWorks() {
        when(repoManager.delete(schema)).thenReturn(true);
        var response = pipelineApi.deleteRepo(schema);

        assertEquals(ResponseEntity.ok().build(), response);
    }

    @Test
    public void testDeleteRepoFails() {
        when(repoManager.delete(schema)).thenReturn(false);
        var response = pipelineApi.deleteRepo(schema);

        assertEquals(ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build(), response);
    }

    @Test
    public void testSyncRepoWorks() {
        when(repoManager.sync(schema)).thenReturn(true);
        var response = pipelineApi.syncRepo(schema);

        assertEquals(ResponseEntity.ok().build(), response);
    }

    @Test
    public void testSyncRepoFails() {
        when(repoManager.sync(schema)).thenReturn(false);
        var response = pipelineApi.syncRepo(schema);

        assertEquals(ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build(), response);
    }
}
