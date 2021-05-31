package com.greenops.workflowtrigger.dbclient;

import com.greenops.workflowtrigger.api.model.pipeline.TeamSchema;

public interface DbClient {

    public boolean store(TeamSchema teamSchema);
    public TeamSchema fetch(String teamName);
}
