package com.greenops.util.dbclient;

import com.fasterxml.jackson.databind.JsonNode;
import com.greenops.util.datamodel.auditlog.DeploymentLog;
import com.greenops.util.datamodel.auditlog.Log;
import com.greenops.util.datamodel.auditlog.RemediationLog;
import com.greenops.util.datamodel.cluster.ClusterSchema;
import com.greenops.util.datamodel.metadata.StepMetadata;
import com.greenops.util.datamodel.metadata.StepMetadata;
import com.greenops.util.datamodel.pipeline.TeamSchema;
import com.greenops.util.datamodel.clientmessages.ClientRequest;

import java.util.List;

public interface DbClient {

    enum ObjectType {
        TEAM_SCHEMA, STRING_LIST, LOG_LIST, SINGLE_LOG, CLIENT_REQUEST, CLUSTER_SCHEMA, METADATA
    }

    enum ListStoreOperation {
        NONE, INSERT, UPDATE;
    }

    static final int LOG_INCREMENT = 15;

    public void storeValue(String key, Object schema);
    public void insertValueInList(String key, Object schema);
    public void insertValueInTransactionlessList(String key, Object schema);
    public void updateHeadInList(String key, Object schema);
    public void updateHeadInTransactionlessList(String key, Object schema);
    public TeamSchema fetchTeamSchema(String key);
    public List<String> fetchStringList(String key);
    public ClusterSchema fetchClusterSchema(String key);
    public ClusterSchema fetchClusterSchemaTransactionless(String key);
    public List<Log> fetchLogList(String key, int increment);
    public Log fetchLatestLog(String key);
    public DeploymentLog fetchLatestDeploymentLog(String key);
    public RemediationLog fetchLatestRemediationLog(String key);
    public StepMetadata fetchMetadata(String key);
    public ClientRequest fetchHeadInClientRequestList(String key);
    public void shutdown();
}

