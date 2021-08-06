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

import static com.greenops.workfloworchestrator.ingest.apiclient.clientwrapper.ClientWrapperApi.DEPLOY_ARGO_REQUEST;
import static com.greenops.workfloworchestrator.ingest.apiclient.clientwrapper.ClientWrapperApi.DEPLOY_TEST_REQUEST;

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
    public void deploy(String clusterName, String orgName, String teamName, String type, Object payload) {
        try {
            var body = type.equals(DEPLOY_TEST_REQUEST) ? objectMapper.writeValueAsString(payload) : (String)payload;
            var deployRequest = new ClientDeployRequest(orgName, type, body);
            var dbKey = DbKey.makeClientRequestQueueKey(orgName, teamName, clusterName);
            dbClient.insertValueInTransactionlessList(dbKey, deployRequest);
        } catch (JsonProcessingException e) {
            log.error("Object mapper could not convert payload to ClientDeployRequest", e);
            throw new AtlasNonRetryableError(e);
        }
    }

    @Override
    public void deployAndWatch(String clusterName, String orgName, String teamName, String pipelineName, String stepName, String deployType, Object payload, String watchType) {
        try {
            var body = deployType.equals(DEPLOY_TEST_REQUEST) ? objectMapper.writeValueAsString(payload) : (String)payload;
            var deployAndWatchRequest = new ClientDeployAndWatchRequest(orgName, deployType, body, watchType, teamName, pipelineName, stepName);
            var dbKey = DbKey.makeClientRequestQueueKey(orgName, teamName, clusterName);
            dbClient.insertValueInTransactionlessList(dbKey, deployAndWatchRequest);
        } catch (JsonProcessingException e) {
            log.error("Object mapper could not convert payload to ClientDeployAndWatchRequest", e);
            throw new AtlasNonRetryableError(e);
        }
    }

    @Override
    public void deployArgoAppByName(String clusterName, String orgName, String teamName, String pipelineName, String stepName, String appName, String watchType) {
        var deployRequest = new ClientDeployNamedArgoApplicationRequest(orgName, DEPLOY_ARGO_REQUEST, appName);
        var dbKey = DbKey.makeClientRequestQueueKey(orgName, teamName, clusterName);
        dbClient.insertValueInTransactionlessList(dbKey, deployRequest);
    }

    @Override
    public void deployArgoAppByNameAndWatch(String clusterName, String orgName, String teamName, String pipelineName, String stepName, String appName, String watchType) {
            var deployAndWatchRequest = new ClientDeployNamedArgoAppAndWatchRequest(orgName, DEPLOY_ARGO_REQUEST, appName, watchType, teamName, pipelineName, stepName);
            var dbKey = DbKey.makeClientRequestQueueKey(orgName, teamName, clusterName);
            dbClient.insertValueInTransactionlessList(dbKey, deployAndWatchRequest);
    }

    @Override
    public void rollbackAndWatch(String clusterName, String orgName, String teamName, String pipelineName, String stepName, String appName, String revisionHash, String watchType) {
        var rollbackAndWatchRequest = new ClientRollbackAndWatchRequest(orgName, appName, revisionHash, watchType, teamName, pipelineName, stepName);
        var dbKey = DbKey.makeClientRequestQueueKey(orgName, teamName, clusterName);
        dbClient.insertValueInTransactionlessList(dbKey, rollbackAndWatchRequest);
    }

    @Override
    public void deleteByConfig(String clusterName, String orgName, String teamName, String type, String configPayload) {
        var deleteRequest = new ClientDeleteByConfigRequest(orgName, type, configPayload);
        var dbKey = DbKey.makeClientRequestQueueKey(orgName, teamName, clusterName);
        dbClient.insertValueInTransactionlessList(dbKey, deleteRequest);
    }

    @Override
    public void deleteByGvk(String clusterName, String orgName, String teamName, String type, String resourceName, String resourceNamespace, String group, String version, String kind) {
        var deleteRequest = new ClientDeleteByGvkRequest(orgName, type, resourceName, resourceNamespace, group, version, kind);
        var dbKey = DbKey.makeClientRequestQueueKey(orgName, teamName, clusterName);
        dbClient.insertValueInTransactionlessList(dbKey, deleteRequest);
    }
}
