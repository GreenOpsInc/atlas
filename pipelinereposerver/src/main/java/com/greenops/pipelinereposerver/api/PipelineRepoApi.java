package com.greenops.pipelinereposerver.api;

import com.greenops.pipelinereposerver.repomanager.RepoManager;
import com.greenops.util.datamodel.git.GitRepoSchema;
import com.greenops.util.datamodel.git.GitRepoSchemaInfo;
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

    @PostMapping(value = "clone/{org}")
    ResponseEntity<Void> cloneRepo(@PathVariable("org") String org, @RequestBody GitRepoSchema gitRepoSchema) {
        if (!repoManager.getOrgName().equals(org)) {
            return ResponseEntity.status(HttpStatus.BAD_REQUEST).build();
        }
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
    ResponseEntity<String> syncRepo(@RequestBody GitRepoSchema gitRepoSchema) {
        var revisionHash = repoManager.sync(gitRepoSchema);
        if (revisionHash != null) {
            return ResponseEntity.ok(revisionHash);
        } else {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }

    @PostMapping(value = "resetToVersion/{orgName}/{teamName}/{gitCommit}")
    ResponseEntity<Void> resetRepoToVersion(@PathVariable("orgName") String orgName,
                                            @PathVariable("teamName") String teamName,
                                            @PathVariable("gitCommit") String gitCommit,
                                            @RequestBody GitRepoSchemaInfo gitRepoSchemaInfo) {
        var gitRepoSchema = new GitRepoSchema(gitRepoSchemaInfo.getGitRepo(), gitRepoSchemaInfo.getPathToRoot(), null);
        if (repoManager.getOrgName().equals(orgName) && repoManager.containsGitRepoSchema(gitRepoSchema)) {
            if (repoManager.resetToVersion(gitCommit, gitRepoSchema)) {
                return ResponseEntity.ok().build();
            } else {
                return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
            }
        } else {
            return ResponseEntity.status(HttpStatus.BAD_REQUEST).build();
        }
    }
}
