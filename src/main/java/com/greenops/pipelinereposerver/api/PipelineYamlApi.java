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
@RequestMapping("/yaml")
public class PipelineYamlApi {

    private final RepoManager repoManager;

    @Autowired
    public PipelineYamlApi(RepoManager repoManager) {
        this.repoManager = repoManager;
    }

    //TODO: This only returns string contents for now, we need to return the actual yaml content as well.
    @GetMapping("/{orgName}/{teamName}")
    public ResponseEntity<String> getPipelineConfig(@PathVariable("orgName") String orgName,
                                                    @PathVariable("teamName") String teamName,
                                                    @RequestBody GetFileRequest fileRequest) {
        if (repoManager.getOrgName().equals(orgName) &&
                repoManager.containsGitRepoSchema(new GitRepoSchema(fileRequest.getGitRepoUrl(), null, null))) {
            String fileContents = repoManager.getYamlFileContents(fileRequest.gitRepoUrl, fileRequest.filename);
            if (fileContents == null) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND).build();
            }
            return ResponseEntity.ok(fileContents);
        }

        return ResponseEntity.status(HttpStatus.NOT_FOUND).build();
    }

    private class GetFileRequest {
        private final String gitRepoUrl;
        private final String filename;

        GetFileRequest(String gitRepoUrl, String filename) {
            this.gitRepoUrl = gitRepoUrl;
            this.filename = filename;
        }

        String getGitRepoUrl() {
            return gitRepoUrl;
        }

        String getFilename() {
            return filename;
        }
    }
}