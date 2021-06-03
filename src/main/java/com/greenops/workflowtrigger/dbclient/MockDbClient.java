package com.greenops.workflowtrigger.dbclient;

import com.greenops.workflowtrigger.api.model.pipeline.TeamSchema;

public class MockDbClient implements DbClient {

    @Override
    public boolean store(String key, TeamSchema teamSchema) {
        return true;
    }

    @Override
    public TeamSchema fetch(String key) {
        return null;
    }
}
