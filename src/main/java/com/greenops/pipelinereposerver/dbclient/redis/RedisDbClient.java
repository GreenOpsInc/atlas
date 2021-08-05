package com.greenops.pipelinereposerver.dbclient.redis;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.pipelinereposerver.api.model.git.GitCredMachineUser;
import com.greenops.pipelinereposerver.api.model.git.GitCredToken;
import com.greenops.pipelinereposerver.api.model.mixin.git.GitCredMachineUserMixin;
import com.greenops.pipelinereposerver.api.model.mixin.git.GitCredTokenMixin;
import com.greenops.pipelinereposerver.api.model.mixin.pipeline.PipelineSchemaMixin;
import com.greenops.pipelinereposerver.api.model.mixin.pipeline.TeamSchemaMixin;
import com.greenops.pipelinereposerver.api.model.pipeline.PipelineSchemaImpl;
import com.greenops.pipelinereposerver.api.model.pipeline.TeamSchema;
import com.greenops.pipelinereposerver.api.model.pipeline.TeamSchemaImpl;
import com.greenops.pipelinereposerver.dbclient.DbClient;
import io.lettuce.core.RedisClient;
import io.lettuce.core.api.StatefulRedisConnection;
import io.lettuce.core.api.sync.RedisCommands;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import javax.annotation.PreDestroy;
import java.util.List;

@Slf4j
@Component
public class RedisDbClient implements DbClient {
    //TODO: Eventually we should have a configuration factory/file which will choose which component to pick. For now this is fine.
    private final RedisClient redisClient;
    private final StatefulRedisConnection<String, String> redisConnection;
    private final RedisCommands<String, String> redisCommands;
    private final ObjectMapper objectMapper;

    @Autowired
    public RedisDbClient(@Value("${application.redis-url}") String redisUrl, ObjectMapper objectMapper) {
        redisClient = RedisClient.create("redis://" + redisUrl); //Pattern is redis://password@host:port
        redisConnection = redisClient.connect();
        redisCommands = redisConnection.sync();
        this.objectMapper =  objectMapper;
    }

    @PreDestroy
    public void shutdown() {
        log.info("Shutting down Redis client...");
        redisConnection.close();
        redisClient.shutdown();
    }

    @Override
    public TeamSchema fetchTeamSchema(String key) {
        return (TeamSchema) fetch(key, ObjectType.TEAM_SCHEMA);
    }

    @Override
    public List<String> fetchList(String key) {
        return (List<String>) fetch(key, ObjectType.LIST);
    }

    private Object fetch(String key, ObjectType objectType) {
        try {
            log.info("Fetching schema for key {}", key);
            var result = redisCommands.get(key);
            //If the key doesn't exist, Redis will return null
            if (result == null) {
                return null;
            } else if (objectType == ObjectType.TEAM_SCHEMA) {
                return objectMapper.readValue(result, TeamSchema.class);
            } else if (objectType == ObjectType.LIST) {
                return objectMapper.readValue(result, objectMapper.getTypeFactory().constructCollectionType(List.class, String.class));
            }
        } catch (JsonProcessingException e) {
            throw new RuntimeException("Jackson object mapping/serialization failed.");
        }
        return null;
    }
}
