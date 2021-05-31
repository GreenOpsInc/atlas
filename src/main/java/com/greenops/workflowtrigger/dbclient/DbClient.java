package com.greenops.workflowtrigger.dbclient;

import com.greenops.workflowtrigger.api.model.git.GitRepoSchema;

public interface DbClient {
    public boolean store(GitRepoSchema gitRepoSchema);
}
