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
        var gitRepoSchema = new GitRepoSchema(fileRequest.getGitRepoSchemaInfo().getGitRepo(), fileRequest.getGitRepoSchemaInfo().getPathToRoot(), null);
        if (repoManager.getOrgName().equals(orgName) &&
                repoManager.containsGitRepoSchema(gitRepoSchema)) {
            //gitCommitHash should never be ROOT_COMMIT
            var desiredGitCommit = fileRequest.getGitCommitHash();
            if (desiredGitCommit.equals(ROOT_COMMIT)) {
                desiredGitCommit = repoManager.getCurrentCommit(gitRepoSchema.getGitRepo());
            }
            if (!repoManager.resetToVersion(desiredGitCommit, gitRepoSchema)) {
                return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body("Could not switch to right revision");
            }
            String fileContents = repoManager.getYamlFileContents(fileRequest.getFilename(), gitRepoSchema);
            if (fileContents == null) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND).body("Couldn't find contents");
            }
            return ResponseEntity.ok(fileContents);
        }

        return ResponseEntity.status(HttpStatus.BAD_REQUEST).build();
    }
}