package com.greenops.workflowtrigger.dbclient;

import com.greenops.workflowtrigger.api.model.pipeline.TeamSchema;

import java.util.List;

public interface DbClient {

    enum ObjectType {
        TEAM_SCHEMA, LIST;
    }

    public boolean store(String key, Object teamSchema);
    public TeamSchema fetchTeamSchema(String key);
    public List<String> fetchList(String key);
}
