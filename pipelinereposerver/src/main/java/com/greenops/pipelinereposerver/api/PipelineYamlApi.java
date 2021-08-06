package com.greenops.pipelinereposerver.api;

import com.greenops.pipelinereposerver.repomanager.RepoManager;
import com.greenops.util.datamodel.git.GitRepoSchema;
import com.greenops.util.datamodel.request.GetFileRequest;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@Slf4j
@RestController
@RequestMapping("/data")
public class PipelineYamlApi {

    public static final String ROOT_COMMIT = "ROOT_COMMIT";

    private final RepoManager repoManager;

    @Autowired
    public PipelineYamlApi(RepoManager repoManager) {
        this.repoManager = repoManager;
    }

    //TODO: This only returns string contents for now, we need to return the actual yaml content as well.
    @PostMapping("/file/{orgName}/{teamName}")
    public ResponseEntity<String> getPipelineConfig(@PathVariable("orgName") String orgName,
                                                    @PathVariable("teamName") String teamName,
                                                    @RequestBody GetFileRequest fileRequest) {
        if (repoManager.getOrgName().equals(orgName) &&
                repoManager.containsGitRepoSchema(new GitRepoSchema(fileRequest.getGitRepoUrl(), null, null))) {
            var desiredGitCommit = fileRequest.getGitCommitHash();
            if (desiredGitCommit.equals(ROOT_COMMIT)) {
                desiredGitCommit = repoManager.getRootCommit(fileRequest.getGitRepoUrl());
            }
            if (!repoManager.getCurrentCommit(fileRequest.getGitRepoUrl()).equals(desiredGitCommit)) {
                var success = repoManager.resetToVersion(desiredGitCommit, fileRequest.getGitRepoUrl());
                if (!success) return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
            }
            String fileContents = repoManager.getYamlFileContents(fileRequest.getGitRepoUrl(), fileRequest.getFilename());
            if (fileContents == null) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND).build();
            }
            return ResponseEntity.ok(fileContents);
        }

        return ResponseEntity.status(HttpStatus.BAD_REQUEST).build();
    }

    @PostMapping("/version/{orgName}/{teamName}")
    public ResponseEntity<String> getCurrentPipelineCommitHash(@PathVariable("orgName") String orgName,
                                                    @PathVariable("teamName") String teamName,
                                                    @RequestBody String gitRepoUrl) {
        if (repoManager.getOrgName().equals(orgName) &&
                repoManager.containsGitRepoSchema(new GitRepoSchema(gitRepoUrl, null, null))) {
            String commitHash = repoManager.getLatestCommitFromCache(gitRepoUrl);
            if (commitHash == null) {
                return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
            }
            return ResponseEntity.ok(commitHash);
        }

        return ResponseEntity.status(HttpStatus.BAD_REQUEST).build();
    }
}