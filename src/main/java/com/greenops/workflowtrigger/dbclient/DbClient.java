package com.greenops.workflowtrigger.dbclient;

import com.greenops.workflowtrigger.api.model.pipeline.TeamSchema;

public interface DbClient {

    public boolean store(String key, TeamSchema teamSchema);
    public TeamSchema fetch(String key);
}
