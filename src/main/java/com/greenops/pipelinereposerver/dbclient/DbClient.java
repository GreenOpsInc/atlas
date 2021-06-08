package com.greenops.pipelinereposerver.dbclient;

import com.greenops.pipelinereposerver.api.model.pipeline.TeamSchema;

import java.util.List;

public interface DbClient {

    enum ObjectType {
        TEAM_SCHEMA, LIST;
    }

    public TeamSchema fetchTeamSchema(String key);
    public List<String> fetchList(String key);
    public void shutdown();

}
