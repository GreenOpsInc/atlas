package com.greenops.workflowtrigger.dbclient.redis;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.workflowtrigger.api.model.git.GitCredMachineUser;
import com.greenops.workflowtrigger.api.model.git.GitRepoSchema;
import com.greenops.workflowtrigger.api.model.mixin.git.GitCredMachineUserMixin;
import com.greenops.workflowtrigger.api.model.mixin.git.GitRepoSchemaMixin;
import com.greenops.workflowtrigger.api.model.mixin.pipeline.PipelineSchemaMixin;
import com.greenops.workflowtrigger.api.model.mixin.pipeline.TeamSchemaMixin;
import com.greenops.workflowtrigger.api.model.pipeline.PipelineSchemaImpl;
import com.greenops.workflowtrigger.api.model.pipeline.TeamSchema;
import com.greenops.workflowtrigger.api.model.pipeline.TeamSchemaImpl;
import com.greenops.workflowtrigger.dbclient.DbClient;
import com.lambdaworks.redis.RedisClient;
import com.lambdaworks.redis.RedisURI;
import org.springframework.stereotype.Component;

@Component
public class RedisDbClient implements DbClient {
    //TODO: Write Redis IT
    private static String REDIS_HOST_VARIABLE = "REDIS_HOST";
    private static String REDIS_PORT_VARIABLE = "REDIS_PORT";
    private static String REDIS_SUCCESS_MESSAGE = "OK";
    //TODO: Eventually we should have a configuration factory/file which will choose which component to pick. For now this is fine.
    private RedisClient client;
    private ObjectMapper objectMapper;

    public RedisDbClient() {
        var host = System.getenv(REDIS_HOST_VARIABLE);
        var port = System.getenv(REDIS_PORT_VARIABLE);
        client = new RedisClient(RedisURI.create("redis://" + host + ":" + port)); //Pattern is redis://password@host:port
        objectMapper = new ObjectMapper()
                .addMixIn(TeamSchemaImpl.class, TeamSchemaMixin.class)
                .addMixIn(PipelineSchemaImpl.class, PipelineSchemaMixin.class)
                .addMixIn(GitRepoSchema.class, GitRepoSchemaMixin.class)
                .addMixIn(GitCredMachineUser.class, GitCredMachineUserMixin.class);
    }

    @Override
    public boolean store(TeamSchema teamSchema) {
        //TODO: The key is wrong. It should be updated when we know how org names are going to be defined
        try {
            var connection= client.connect();
            var result = connection.set(teamSchema.getTeamName(), objectMapper.writeValueAsString(teamSchema));
            connection.close();
            return result.equals(REDIS_SUCCESS_MESSAGE);
        } catch (JsonProcessingException e) {
            throw new RuntimeException("Jackson object mapping/serialization failed.");
        }
    }

    @Override
    public TeamSchema fetch(String teamName) {
        try {
            var connection= client.connect();
            //TODO: The key is wrong. It should be updated when we know how org names are going to be defined
            var result = connection.get(teamName);
            connection.close();
            return objectMapper.readValue(result, TeamSchema.class);
        } catch (JsonProcessingException e) {
            throw new RuntimeException("Jackson object mapping/serialization failed.");
        }
    }
}
