package com.greenops.workflowtrigger.dbclient;

import com.greenops.workflowtrigger.api.model.GitRepoSchema;

public interface DbClient {
    public boolean store(GitRepoSchema gitRepoSchema);
}
