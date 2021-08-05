package com.greenops.workflowtrigger.dbclient;

import com.greenops.workflowtrigger.api.model.auditlog.DeploymentLog;
import com.greenops.workflowtrigger.api.model.cluster.ClusterSchema;
import com.greenops.workflowtrigger.api.model.pipeline.TeamSchema;

import java.util.List;

public interface DbClient {

    enum ObjectType {
        TEAM_SCHEMA, STRING_LIST, LOG_LIST, SINGLE_LOG, CLUSTER_SCHEMA;
    }

    enum ListStoreOperation {
        NONE, INSERT, UPDATE;
    }

    static final int LOG_INCREMENT = 15;

    public void storeValue(String key, Object schema);
    public void insertValueInList(String key, Object schema);
    public void updateHeadInList(String key, Object schema);
    public TeamSchema fetchTeamSchema(String key);
    public List<String> fetchStringList(String key);
    public ClusterSchema fetchClusterSchema(String key);
    public List<DeploymentLog> fetchLogList(String key, int increment);
    public DeploymentLog fetchLatestLog(String key);
}

