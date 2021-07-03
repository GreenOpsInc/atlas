package com.greenops.workfloworchestrator.ingest.dbclient;

import com.greenops.workfloworchestrator.datamodel.auditlog.DeploymentLog;
import com.greenops.workfloworchestrator.datamodel.pipelineschema.TeamSchema;

import java.util.List;

public interface DbClient {

    enum ObjectType {
        TEAM_SCHEMA, STRING_LIST, LOG_LIST, SINGLE_LOG;
    }

    enum ListStoreOperation {
        NONE, INSERT, UPDATE;
    }

    public boolean storeValue(String key, Object schema);
    public boolean insertValueInList(String key, Object schema);
    public boolean updateHeadInList(String key, Object schema);
    public TeamSchema fetchTeamSchema(String key);
    public List<String> fetchStringList(String key);
    public List<DeploymentLog> fetchLogList(String key);
    public DeploymentLog fetchLatestLog(String key);
}
