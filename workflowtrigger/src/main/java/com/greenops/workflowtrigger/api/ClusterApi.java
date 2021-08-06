package com.greenops.workflowtrigger.api;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.util.datamodel.cluster.ClusterSchema;
import com.greenops.util.dbclient.DbClient;
import com.greenops.workflowtrigger.dbclient.DbKey;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@Slf4j
@RestController
@RequestMapping("/cluster")
public class ClusterApi {
    private final DbClient dbClient;
    private final ObjectMapper objectMapper;

    @Autowired
    public ClusterApi(DbClient dbClient, ObjectMapper objectMapper) {
        this.dbClient = dbClient;
        this.objectMapper = objectMapper;
    }

    @PostMapping(value = "/{orgName}")
    public ResponseEntity<Void> createCluster(@PathVariable("orgName") String orgName,
                                              @RequestBody ClusterSchema clusterSchema) {
        var key = DbKey.makeDbClusterKey(orgName, clusterSchema.getClusterName());

        if (dbClient.fetchClusterSchema(key) != null) {
            return ResponseEntity.status(HttpStatus.CONFLICT).build();
        }

        dbClient.storeValue(key, clusterSchema);
        return ResponseEntity.ok().build();

    }

    @GetMapping(value = "/{orgName}/{clusterName}")
    public ResponseEntity<String> readCluster(@PathVariable("orgName") String orgName,
                                              @PathVariable("clusterName") String clusterName) {
        var key = DbKey.makeDbClusterKey(orgName, clusterName);
        var clusterSchema = dbClient.fetchClusterSchema(key);
        if (clusterSchema == null) {
            return ResponseEntity.badRequest().build();
        }

        return ResponseEntity.ok()
                .contentType(MediaType.APPLICATION_JSON)
                .body(schemaToResponsePayload(clusterSchema));

    }

    @DeleteMapping(value = "/{orgName}/{clusterName}")
    public ResponseEntity<Void> deleteCluster(@PathVariable("orgName") String orgName,
                                              @PathVariable("clusterName") String clusterName) {
        var key = DbKey.makeDbClusterKey(orgName, clusterName);
        dbClient.storeValue(key, null);
        return ResponseEntity.ok().build();
    }

    private String schemaToResponsePayload(Object schema) {
        try {
            return objectMapper.writeValueAsString(schema);
        } catch (JsonProcessingException e) {
            throw new RuntimeException("Could not convert schema into response payload.", e);
        }
    }

}

