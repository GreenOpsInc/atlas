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
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

@Slf4j
@Component
public class RedisDbClient implements DbClient {
    //TODO: Write Redis IT
    private static final String REDIS_SUCCESS_MESSAGE = "OK";
    //TODO: Eventually we should have a configuration factory/file which will choose which component to pick. For now this is fine.
    private RedisClient client;
    private ObjectMapper objectMapper;

    public RedisDbClient(@Value("${application.redis-url}") String redisUrl) {
        client = new RedisClient(RedisURI.create("redis://" + redisUrl)); //Pattern is redis://password@host:port
        objectMapper = new ObjectMapper()
                .addMixIn(TeamSchemaImpl.class, TeamSchemaMixin.class)
                .addMixIn(PipelineSchemaImpl.class, PipelineSchemaMixin.class)
                .addMixIn(GitRepoSchema.class, GitRepoSchemaMixin.class)
                .addMixIn(GitCredMachineUser.class, GitCredMachineUserMixin.class);
    }

    @Override
    public boolean store(String key, TeamSchema teamSchema) {
        try {
            log.info("Storing schema for team {}", teamSchema.getTeamName());
            var connection= client.connect();
            var result = connection.set(key, objectMapper.writeValueAsString(teamSchema));
            connection.close();
            return result.equals(REDIS_SUCCESS_MESSAGE);
        } catch (JsonProcessingException e) {
            throw new RuntimeException("Jackson object mapping/serialization failed.");
        }
    }

    @Override
    public TeamSchema fetch(String key) {
        try {
            log.info("Fetching schema for team {}", key);
            var connection= client.connect();
            var result = connection.get(key);
            connection.close();
            //If the key doesn't exist, Redis will return null
            return result != null ? objectMapper.readValue(result, TeamSchema.class) : null;
        } catch (JsonProcessingException e) {
            throw new RuntimeException("Jackson object mapping/serialization failed.");
        }
    }
}
