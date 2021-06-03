package com.greenops.workflowtrigger.api;

import com.greenops.workflowtrigger.api.model.git.GitRepoSchema;
import com.greenops.workflowtrigger.api.reposerver.RepoManagerApi;
import com.greenops.workflowtrigger.dbclient.DbClient;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@Slf4j
@RestController
@RequestMapping("/pipeline")
public class PipelineApi {

    @Autowired
    private DbClient dbClient;

    @Autowired
    private RepoManagerApi repoManagerApi;

    @PostMapping(value = "/{teamName}/{pipelineName}")
    public ResponseEntity<Void> createPipeline(@PathVariable("teamName") String teamName,
                                               @PathVariable("pipelineName") String pipelineName,
                                               @RequestBody GitRepoSchema gitRepo) {
        // TODO: implement pipeline creation logic
        return ResponseEntity.ok().build();
    }

    @GetMapping(value = "/{teamName}/{pipelineName}")
    public ResponseEntity<GitRepoSchema> getPipeline(@PathVariable("teamName") String teamName, @PathVariable("pipelineName") String pipelineName) {
        // TODO: implement pipeline fetch logic
        return ResponseEntity.ok().build();
    }

    @PutMapping(value = "/{teamName}/{pipelineName}")
    public ResponseEntity<Void> updatePipeline(@PathVariable("teamName") String teamName, @PathVariable("pipelineName") String pipelineName, @RequestBody(required = false) GitRepoSchema gitRepo) {
        // TODO: implement pipeline update logic
        return ResponseEntity.ok().build();
    }

    @DeleteMapping(value = "/{teamName}/{pipelineName}")
    public ResponseEntity<Void> deletePipeline(@PathVariable("teamName") String teamName, @PathVariable("pipelineName") String pipelineName, @RequestBody GitRepoSchema gitRepo) {
        // TODO: implement pipeline deletion logic
        return ResponseEntity.ok().build();
    }

    public PipelineApi(DbClient dbClient) {
        this.dbClient = dbClient;
    }
}
