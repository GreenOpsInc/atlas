package com.greenops.workfloworchestrator.ingest.apiclient.clientwrapper;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.util.datamodel.clientmessages.*;
import com.greenops.util.dbclient.DbClient;
import com.greenops.workfloworchestrator.error.AtlasNonRetryableError;
import com.greenops.workfloworchestrator.ingest.dbclient.DbKey;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Component;

import static com.greenops.workfloworchestrator.ingest.apiclient.clientwrapper.ClientRequestQueue.DEPLOY_ARGO_REQUEST;
import static com.greenops.workfloworchestrator.ingest.apiclient.clientwrapper.ClientRequestQueue.DEPLOY_TEST_REQUEST;

@Slf4j
@Component
public class ClientRequestQueueImpl implements ClientRequestQueue {

    private final DbClient dbClient;
    private final ObjectMapper objectMapper;

    @Autowired
    public ClientRequestQueueImpl(DbClient dbClient, @Qualifier("eventAndRequestObjectMapper") ObjectMapper objectMapper) {
        this.dbClient = dbClient;
        this.objectMapper = objectMapper;
    }

    @Override
    public void deploy(String clusterName, String orgName, String teamName, String pipelineName, String uvn, String stepName, String responseEventType, String type, String revisionHash, Object payload) {
        try {
            var body = type.equals(DEPLOY_TEST_REQUEST) ? objectMapper.writeValueAsString(payload) : (String)payload;
            var deployRequest = new ClientDeployRequest(orgName, teamName, pipelineName, uvn, stepName, responseEventType, type, revisionHash, body);
            var dbKey = DbKey.makeClientRequestQueueKey(orgName, clusterName);
            dbClient.insertValueInTransactionlessList(dbKey, deployRequest);
        } catch (JsonProcessingException e) {
            log.error("Object mapper could not convert payload to ClientDeployRequest", e);
            throw new AtlasNonRetryableError(e);
        }
    }

    @Override
    public void deployAndWatch(String clusterName, String orgName, String teamName, String pipelineName, String uvn, String stepName, String deployType, String revisionHash, Object payload, String watchType, int testNumber) {
        try {
            var body = deployType.equals(DEPLOY_TEST_REQUEST) ? objectMapper.writeValueAsString(payload) : (String)payload;
            var deployAndWatchRequest = new ClientDeployAndWatchRequest(orgName, uvn, deployType, revisionHash, body, watchType, teamName, pipelineName, stepName, testNumber);
            var dbKey = DbKey.makeClientRequestQueueKey(orgName, clusterName);
            dbClient.insertValueInTransactionlessList(dbKey, deployAndWatchRequest);
        } catch (JsonProcessingException e) {
            log.error("Object mapper could not convert payload to ClientDeployAndWatchRequest", e);
            throw new AtlasNonRetryableError(e);
        }
    }

    @Override
    public void selectiveSyncArgoApplication(String clusterName, String orgName, String teamName, String pipelineName, String uvn, String stepName, String revisionHash, ResourcesGvkRequest resourcesGvkRequest, String appName) {
        var deployRequest = new ClientSelectiveSyncAndWatchRequest(orgName, teamName, pipelineName, uvn, stepName, revisionHash, appName, resourcesGvkRequest);
        var dbKey = DbKey.makeClientRequestQueueKey(orgName, clusterName);
        dbClient.insertValueInTransactionlessList(dbKey, deployRequest);
    }

    @Override
    public void deployArgoAppByName(String clusterName, String orgName, String pipelineName, String stepName, String appName, String watchType) {
        var deployRequest = new ClientDeployNamedArgoApplicationRequest(orgName, DEPLOY_ARGO_REQUEST, appName);
        var dbKey = DbKey.makeClientRequestQueueKey(orgName, clusterName);
        dbClient.insertValueInTransactionlessList(dbKey, deployRequest);
    }

    @Override
    public void deployArgoAppByNameAndWatch(String clusterName, String orgName, String teamName, String pipelineName, String uvn, String stepName, String appName, String watchType) {
            var deployAndWatchRequest = new ClientDeployNamedArgoAppAndWatchRequest(orgName, uvn, DEPLOY_ARGO_REQUEST, appName, watchType, teamName, pipelineName, stepName);
            var dbKey = DbKey.makeClientRequestQueueKey(orgName, clusterName);
            dbClient.insertValueInTransactionlessList(dbKey, deployAndWatchRequest);
    }

    @Override
    public void rollbackAndWatch(String clusterName, String orgName, String teamName, String pipelineName, String uvn, String stepName, String appName, String revisionHash, String watchType) {
        var rollbackAndWatchRequest = new ClientRollbackAndWatchRequest(orgName, uvn, appName, revisionHash, watchType, teamName, pipelineName, stepName);
        var dbKey = DbKey.makeClientRequestQueueKey(orgName, clusterName);
        dbClient.insertValueInTransactionlessList(dbKey, rollbackAndWatchRequest);
    }

    @Override
    public void deleteByConfig(String clusterName, String orgName, String teamName, String pipelineName, String uvn, String stepName, String type, String configPayload) {
        var deleteRequest = new ClientDeleteByConfigRequest(orgName, teamName, pipelineName, uvn, stepName, type, configPayload);
        var dbKey = DbKey.makeClientRequestQueueKey(orgName, clusterName);
        dbClient.insertValueInTransactionlessList(dbKey, deleteRequest);
    }

    @Override
    public void deleteByGvk(String clusterName, String orgName, String teamName, String pipelineName, String uvn, String stepName, String type, String resourceName, String resourceNamespace, String group, String version, String kind) {
        var deleteRequest = new ClientDeleteByGvkRequest(orgName, teamName, pipelineName, uvn, stepName, type, resourceName, resourceNamespace, group, version, kind);
        var dbKey = DbKey.makeClientRequestQueueKey(orgName, clusterName);
        dbClient.insertValueInTransactionlessList(dbKey, deleteRequest);
    }
}
