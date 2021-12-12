package com.greenops.commanddelegator.api;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.util.datamodel.clientmessages.ClientRequest;
import com.greenops.util.datamodel.clientmessages.ClientRequestPacket;
import com.greenops.util.datamodel.clientmessages.NotificationRequest;
import com.greenops.util.dbclient.DbClient;
import com.greenops.util.error.AtlasBadKeyError;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.UUID;

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
        var noDeployInfo = clusterSchema.getNoDeployInfo();
        var notificationQueueKey = DbKey.makeClientNotificationQueueKey(orgName, clusterName);
        var requestQueueKey = DbKey.makeClientRequestQueueKey(orgName, clusterName);
        try {
            //Notifications are from the management plane, they take priority over basic processing
            //Notifications also do not change the state of deployments and clusters, so can be processed even when in a no-deploy state
            var clientNotificationHead = dbClient.fetchHeadInClientRequestList(notificationQueueKey);
            if (clientNotificationHead != null) {
                //Notifications don't subscribe to the retry/final try logic, the client wrapper makes sure it's only run once
                return ResponseEntity.ok().body(List.of(clientNotificationHead.getClientRequest()));
            }
            if (noDeployInfo != null && noDeployInfo.getNamespace().isEmpty()) {
                return ResponseEntity.ok().build();
            }
            var clientRequestPacket = dbClient.fetchHeadInClientRequestList(requestQueueKey);
            if (clientRequestPacket != null) {
                if (noDeployInfo != null && clientRequestPacket.getNamespace().equals(noDeployInfo.getNamespace())) {
                    return ResponseEntity.ok().build();
                }
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
    public ResponseEntity<String> addNotificationCommand(@PathVariable("orgName") String orgName,
                                                         @PathVariable("clusterName") String clusterName,
                                                         @RequestBody ClientRequest notification) {
        var clusterKey = DbKey.makeDbClusterKey(orgName, clusterName);
        var clusterSchema = dbClient.fetchClusterSchemaTransactionless(clusterKey);
        if (clusterSchema == null) {
            return ResponseEntity.badRequest().build();
        }
        var key = DbKey.makeClientNotificationQueueKey(orgName, clusterName);
        var requestId = UUID.randomUUID().toString();
        try {
            ((NotificationRequest) notification).setRequestId(requestId);
            dbClient.insertValueInTransactionlessList(key, new ClientRequestPacket("", notification));
        } catch (AtlasBadKeyError err) {
            return ResponseEntity.badRequest().build();
        }
        return ResponseEntity.ok().body(requestId);
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
        if (dbClient.fetchClusterSchemaTransactionless(clusterKey) == null) {
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