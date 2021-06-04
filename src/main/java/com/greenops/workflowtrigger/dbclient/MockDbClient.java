package com.greenops.workflowtrigger.dbclient;

import com.greenops.workflowtrigger.api.model.pipeline.TeamSchema;

import java.util.List;

public class MockDbClient implements DbClient {

    @Override
    public boolean store(String key, Object teamSchema) {
        return true;
    }

    @Override
    public TeamSchema fetchTeamSchema(String key) {
        return null;
    }

    @Override
    public List<String> fetchList(String key) {
        return null;
    }
}
