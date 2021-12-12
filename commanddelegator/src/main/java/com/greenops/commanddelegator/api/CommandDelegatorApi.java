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

    private static final String LOCAL_CLUSTER_NAME = "kubernetes_local";

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
        var clusterSchema = dbClient.fetchClusterSchemaTransactionless(clusterKey);
        if (clusterSchema == null) {
            return ResponseEntity.badRequest().build();
        }
        //If no deploy is enabled for the cluster, respond with an empty payload
        if (clusterSchema.getNoDeployInfo() == null) {
            return ResponseEntity.ok().build();
        }
        var notificationQueueKey = DbKey.makeClientNotificationQueueKey(orgName, clusterName);
        var requestQueueKey = DbKey.makeClientRequestQueueKey(orgName, clusterName);
        try {
            //Notifications are from the management plane, they take priority over basic processing
            var clientNotificationHead = dbClient.fetchHeadInClientRequestList(notificationQueueKey);
            if (clientNotificationHead != null) {
                return ResponseEntity.ok().body(List.of(clientNotificationHead));
            }
            var clientRequestPacket = dbClient.fetchHeadInClientRequestList(requestQueueKey);
            if (clientRequestPacket != null) {
                var request = clientRequestPacket.getClientRequest();
                request.setFinalTry(clientRequestPacket.getRetryCount() >= 5);
                return ResponseEntity.ok().body(List.of(request));
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

    @PostMapping(value = "/notifications/{orgName}/{clusterName}")
    public ResponseEntity<Void> addNotificationCommand(@PathVariable("orgName") String orgName,
                                                       @PathVariable("clusterName") String clusterName,
                                                       @RequestBody ClientRequest notification) {
        var clusterKey = DbKey.makeDbClusterKey(orgName, clusterName);
        var clusterSchema = dbClient.fetchClusterSchemaTransactionless(clusterKey);
        if (clusterSchema == null) {
            return ResponseEntity.badRequest().build();
        }
        //If no deploy is enabled for the cluster, respond with an empty payload
        if (clusterSchema.getNoDeployInfo() == null) {
            return ResponseEntity.ok().build();
        }
        var key = DbKey.makeClientNotificationQueueKey(orgName, clusterName);
        try {
            dbClient.insertValueInTransactionlessList(key, notification);
        } catch (AtlasBadKeyError err) {
            return ResponseEntity.badRequest().build();
        }
        return ResponseEntity.ok().build();
    }

    @DeleteMapping(value = "/notifications/ackHead/{orgName}/{clusterName}")
    public ResponseEntity<Void> ackHeadOfNotificationList(@PathVariable("orgName") String orgName,
                                                          @PathVariable("clusterName") String clusterName) {
        var clusterKey = DbKey.makeDbClusterKey(orgName, clusterName);
        if (dbClient.fetchClusterSchemaTransactionless(clusterKey) == null) {
            return ResponseEntity.badRequest().build();
        }
        var key = DbKey.makeClientNotificationQueueKey(orgName, clusterName);
        try {
            dbClient.updateHeadInTransactionlessList(key, null);
        } catch (AtlasBadKeyError err) {
            return ResponseEntity.badRequest().build();
        }
        return ResponseEntity.ok().build();
    }

    @DeleteMapping(value = "/requests/retry/{orgName}/{clusterName}")
    public ResponseEntity<Void> retryMessage(@PathVariable("orgName") String orgName,
                                             @PathVariable("clusterName") String clusterName) {
        var clusterKey = DbKey.makeDbClusterKey(orgName, clusterName);
        if (!clusterName.equals(LOCAL_CLUSTER_NAME) && dbClient.fetchClusterSchemaTransactionless(clusterKey) == null) {
            return ResponseEntity.badRequest().build();
        }
        var key = DbKey.makeClientRequestQueueKey(orgName, clusterName);
        try {
            var clientRequestPacket = dbClient.fetchHeadInClientRequestList(key);
            if (clientRequestPacket != null) {
                clientRequestPacket.setRetryCount(clientRequestPacket.getRetryCount() + 1);
            }
            dbClient.insertValueInTransactionlessList(key, clientRequestPacket);
            dbClient.updateHeadInTransactionlessList(key, null);
        } catch (AtlasBadKeyError err) {
            return ResponseEntity.badRequest().build();
        }
        return ResponseEntity.ok().build();
    }
}