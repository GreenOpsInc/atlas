package com.greenops.commanddelegator.api;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.util.datamodel.clientmessages.ClientRequest;
import com.greenops.util.dbclient.DbClient;
import com.greenops.util.error.AtlasBadKeyError;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@Slf4j
@RestController
@RequestMapping("/")
public class CommandDelegatorApi {

    private final DbClient dbClient;
    private final ObjectMapper objectMapper;

    @Autowired
    public CommandDelegatorApi(DbClient dbClient, ObjectMapper objectMapper) {
        this.dbClient = dbClient;
        this.objectMapper = objectMapper;
    }

    @GetMapping(value = "/requests/{orgName}/{clusterName}")
    public ResponseEntity<List<ClientRequest>> getCommands(@PathVariable("orgName") String orgName, @PathVariable("clusterName") String clusterName) {
        var clusterKey = DbKey.makeDbClusterKey(orgName, clusterName);
        if (dbClient.fetchClusterSchemaTransactionless(clusterKey) == null) {
            return ResponseEntity.badRequest().build();
        }
        var key = DbKey.makeClientRequestQueueKey(orgName, clusterName);
        try {
            var clientRequestHead = dbClient.fetchHeadInClientRequestList(key);
            if (clientRequestHead != null) {
                return ResponseEntity.ok().body(List.of(clientRequestHead));
            }
        } catch (AtlasBadKeyError err) {
            return ResponseEntity.badRequest().build();
        }
        return ResponseEntity.ok().build();
    }

    @DeleteMapping(value = "/requests/ackHead/{orgName}/{clusterName}")
    public ResponseEntity<Void> ackHeadOfRequestList(@PathVariable("orgName") String orgName,
                                           @PathVariable("clusterName") String clusterName) {
        var clusterKey = DbKey.makeDbClusterKey(orgName, clusterName);
        if (dbClient.fetchClusterSchemaTransactionless(clusterKey) == null) {
            return ResponseEntity.badRequest().build();
        }
        var key = DbKey.makeClientRequestQueueKey(orgName, clusterName);
        try {
            dbClient.updateHeadInTransactionlessList(key, null);
        } catch (AtlasBadKeyError err) {
            return ResponseEntity.badRequest().build();
        }
        return ResponseEntity.ok().build();
    }
}