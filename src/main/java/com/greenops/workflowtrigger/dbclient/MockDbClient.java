package com.greenops.workflowtrigger.dbclient;

import com.greenops.workflowtrigger.api.model.git.GitRepoSchema;

public class MockDbClient implements DbClient {
    @Override
    public boolean store(GitRepoSchema gitRepoSchema) {
        return true;
    }
}
