package com.greenops.pipelinereposerver.api;

import com.greenops.pipelinereposerver.api.model.git.GitRepoSchema;
import com.greenops.pipelinereposerver.repomanager.RepoManager;
import com.greenops.pipelinereposerver.repomanager.RepoManagerImpl;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/repo")
public class PipelineRepoApi {
    private final String orgName = "Temporary"; //TODO: Needs to be updated when we decide how to configure organization
    private final RepoManager repoManager = new RepoManagerImpl();

    @PostMapping(value = "clone")
    ResponseEntity<Void> cloneRepo(@RequestBody GitRepoSchema gitRepoSchema) {
        if (repoManager.clone(gitRepoSchema)) {
            return ResponseEntity.ok().build();
        } else {
            return ResponseEntity.status(HttpStatus.FAILED_DEPENDENCY).build();
        }
    }

    @DeleteMapping(value = "delete")
    ResponseEntity<Void> deleteRepo(@RequestBody GitRepoSchema gitRepoSchema) {
        if (repoManager.delete(gitRepoSchema)) {
            return ResponseEntity.ok().build();
        } else {
            return ResponseEntity.status(HttpStatus.FAILED_DEPENDENCY).build();
        }
    }
}
