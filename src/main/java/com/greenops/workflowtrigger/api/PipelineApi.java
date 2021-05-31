package com.greenops.workflowtrigger.api;

import com.greenops.workflowtrigger.api.model.git.GitRepoSchema;
import com.greenops.workflowtrigger.dbclient.DbClient;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/pipeline")
public class PipelineApi {

    @Autowired
    private DbClient dbClient;

    @PostMapping()
    public ResponseEntity<Void> createPipeline(@RequestBody GitRepoSchema gitRepo) {
        // TODO: implement pipeline creation logic
        dbClient.store(gitRepo);
        return ResponseEntity.ok().build();
    }

    @GetMapping(value = "{name}")
    public ResponseEntity<GitRepoSchema> getPipeline(@PathVariable("name") String name) {
        // TODO: implement pipeline fetch logic
        return ResponseEntity.ok().build();
    }

    @PutMapping(value = "{name}")
    public ResponseEntity<Void> updatePipeline(@PathVariable("name") String name, @RequestBody(required = false) GitRepoSchema gitRepo) {
        // TODO: implement pipeline update logic
        return ResponseEntity.ok().build();
    }

    @DeleteMapping(value = "{name}")
    public ResponseEntity<Void> deletePipeline(@PathVariable("name") String name, @RequestBody GitRepoSchema gitRepo) {
        // TODO: implement pipeline deletion logic
        return ResponseEntity.ok().build();
    }

    public PipelineApi(DbClient dbClient) {
        this.dbClient = dbClient;
    }
}
