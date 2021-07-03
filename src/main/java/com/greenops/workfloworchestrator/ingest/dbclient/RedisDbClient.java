package com.greenops.workfloworchestrator.ingest.dbclient;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.workfloworchestrator.datamodel.auditlog.DeploymentLog;
import com.greenops.workfloworchestrator.datamodel.pipelineschema.TeamSchema;
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
    //TODO: Write Redis IT
    private static final String REDIS_SUCCESS_MESSAGE = "OK";
    //TODO: Eventually we should have a configuration factory/file which will choose which component to pick. For now this is fine.
    private final RedisClient redisClient;
    private final StatefulRedisConnection<String, String> redisConnection;
    private final RedisCommands<String, String> redisCommands;
    private final ObjectMapper objectMapper;
    private String currentWatchedKey;

    @Autowired
    public RedisDbClient(@Value("${application.redis-url}") String redisUrl, ObjectMapper objectMapper) {
        redisClient = RedisClient.create("redis://" + redisUrl); //Pattern is redis://password@host:port
        redisConnection = redisClient.connect();
        redisCommands = redisConnection.sync();
        this.objectMapper = objectMapper;
    }

    @PreDestroy
    public void destroy() {
        log.info("Shutting down Redis client...");
        redisCommands.unwatch();
        redisConnection.close();
        redisClient.shutdown();
    }

    @Override
    public boolean storeValue(String key, Object schema) {
        return store(key, schema, ListStoreOperation.NONE);
    }

    @Override
    public boolean insertValueInList(String key, Object schema) {
        return store(key, schema, ListStoreOperation.INSERT);
    }

    @Override
    public boolean updateHeadInList(String key, Object schema) {
        return store(key, schema, ListStoreOperation.UPDATE);
    }

    private boolean store(String key, Object schema, ListStoreOperation listStoreOperation) {
        try {
            log.info("Storing schema for key {}", key);
            if (!key.equals(currentWatchedKey)) {
                redisCommands.unwatch();
                currentWatchedKey = null;
            }
            redisCommands.multi();
            //Passing in a null means the key should be deleted
            if (listStoreOperation == ListStoreOperation.NONE) {
                if (schema == null) {
                    redisCommands.del(key);
                } else {
                    redisCommands.set(key, objectMapper.writeValueAsString(schema));
                }
            } else if (listStoreOperation == ListStoreOperation.INSERT) {
                if (schema == null) {
                    redisCommands.lpop(key);
                } else {
                    redisCommands.lpush(key, objectMapper.writeValueAsString(schema));
                }
            } else if (listStoreOperation == ListStoreOperation.UPDATE) {
                if (schema == null) {
                    redisCommands.lpop(key);
                } else {
                    redisCommands.lset(key, 0, objectMapper.writeValueAsString(schema));
                }
            }
            var result = redisCommands.exec();
            //Either all of the transaction is processed or all of it is discarded
            return !result.wasDiscarded();
        } catch (JsonProcessingException e) {
            throw new RuntimeException("Jackson object mapping/serialization failed.");
        }
    }

    @Override
    public TeamSchema fetchTeamSchema(String key) {
        return (TeamSchema) fetch(key, ObjectType.TEAM_SCHEMA);
    }

    @Override
    public List<String> fetchStringList(String key) {
        return (List<String>) fetch(key, ObjectType.STRING_LIST);
    }

    @Override
    public List<DeploymentLog> fetchLogList(String key) {
        return (List<DeploymentLog>) fetch(key, ObjectType.LOG_LIST);
    }

    @Override
    public DeploymentLog fetchLatestLog(String key) {
        return (DeploymentLog) fetch(key, ObjectType.SINGLE_LOG);
    }

    private Object fetch(String key, ObjectType objectType) {
        try {
            log.info("Fetching schema for key {}", key);
            //This will ensure that the list of watched keys is kept to one, so no leakage will happen.
            redisCommands.unwatch();
            redisCommands.watch(key);
            currentWatchedKey = key;
            var result = objectType == ObjectType.SINGLE_LOG ? redisCommands.lindex(key, 0) : redisCommands.get(key);
            //If the key doesn't exist, Redis will return null
            if (result == null) {
                return null;
            } else if (objectType == ObjectType.TEAM_SCHEMA) {
                return objectMapper.readValue(result, TeamSchema.class);
            } else if (objectType == ObjectType.STRING_LIST) {
                return objectMapper.readValue(result, objectMapper.getTypeFactory().constructCollectionType(List.class, String.class));
            } else if (objectType == ObjectType.LOG_LIST) {
                return objectMapper.readValue(result, objectMapper.getTypeFactory().constructCollectionType(List.class, DeploymentLog.class));
            } else if (objectType == ObjectType.SINGLE_LOG) {
                return objectMapper.readValue(result, DeploymentLog.class);
            }
        } catch (JsonProcessingException e) {
            throw new RuntimeException("Jackson object mapping/serialization failed.");
        }
        return null;
    }
}
