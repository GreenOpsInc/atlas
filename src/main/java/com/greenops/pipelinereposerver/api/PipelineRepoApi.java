package com.greenops.pipelinereposerver.api;

import com.greenops.pipelinereposerver.api.model.git.GitRepoSchema;
import com.greenops.pipelinereposerver.repomanager.RepoManager;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@Slf4j
@RestController
@RequestMapping("/repo")
public class PipelineRepoApi {

    private final RepoManager repoManager;

    @Autowired
    public PipelineRepoApi(RepoManager repoManager) {
        this.repoManager = repoManager;
    }

    @PostMapping()
    ResponseEntity<Void> cloneRepo(@RequestBody GitRepoSchema gitRepoSchema) {
        if (repoManager.containsGitRepoSchema(gitRepoSchema)) {
            if (repoManager.update(gitRepoSchema)) {
                return ResponseEntity.ok().build();
            } else {
                return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
            }
        } else {
            if (repoManager.clone(gitRepoSchema)) {
                return ResponseEntity.ok().build();
            } else {
                return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
            }
        }
    }

    @PostMapping(value = "delete")
    ResponseEntity<Void> deleteRepo(@RequestBody GitRepoSchema gitRepoSchema) {
        if (repoManager.delete(gitRepoSchema)) {
            return ResponseEntity.ok().build();
        } else {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }

    @PostMapping(value = "sync")
    ResponseEntity<Void> syncRepo(@RequestBody GitRepoSchema gitRepoSchema) {
        if (repoManager.sync(gitRepoSchema)) {
            return ResponseEntity.ok().build();
        } else {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }
}
