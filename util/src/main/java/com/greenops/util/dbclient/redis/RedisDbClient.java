package com.greenops.util.dbclient.redis;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.util.datamodel.auditlog.DeploymentLog;
import com.greenops.util.datamodel.auditlog.Log;
import com.greenops.util.datamodel.auditlog.PipelineInfo;
import com.greenops.util.datamodel.auditlog.RemediationLog;
import com.greenops.util.datamodel.clientmessages.ClientRequestPacket;
import com.greenops.util.datamodel.cluster.ClusterSchema;
import com.greenops.util.datamodel.metadata.StepMetadata;
import com.greenops.util.datamodel.pipeline.TeamSchema;
import com.greenops.util.datamodel.clientmessages.ClientRequest;
import com.greenops.util.dbclient.DbClient;
import com.greenops.util.error.AtlasBadKeyError;
import com.greenops.util.error.AtlasNonRetryableError;
import com.greenops.util.error.AtlasRetryableError;
import io.lettuce.core.RedisClient;
import io.lettuce.core.api.StatefulRedisConnection;
import io.lettuce.core.api.sync.RedisCommands;
import lombok.extern.slf4j.Slf4j;

import javax.annotation.PreDestroy;
import java.util.ArrayList;
import java.util.List;

@Slf4j
public class RedisDbClient implements DbClient {
    private static final String REDIS_SUCCESS_MESSAGE = "OK";
    //TODO: Eventually we should have a configuration factory/file which will choose which component to pick. For now this is fine.
    private final RedisClient redisClient;
    private final StatefulRedisConnection<String, String> redisConnection;
    private final RedisCommands<String, String> redisCommands;
    private final ObjectMapper objectMapper;
    private String currentWatchedKey;

    public RedisDbClient(String redisUrl, ObjectMapper objectMapper) {
        redisClient = RedisClient.create("redis://" + redisUrl); //Pattern is redis://password@host:port
        redisConnection = redisClient.connect();
        redisCommands = redisConnection.sync();
        this.objectMapper = objectMapper;
    }

    @Override
    @PreDestroy
    public void shutdown() {
        log.info("Shutting down Redis client...");
        redisCommands.unwatch();
        redisConnection.close();
        redisClient.shutdown();
    }

    @Override
    public void storeValue(String key, Object schema) {
        store(key, schema, ListStoreOperation.NONE);
    }

    @Override
    public void insertValueInList(String key, Object schema) {
        store(key, schema, ListStoreOperation.INSERT);
    }

    //This is done explicitly due to how rare the use case is. Should generally never be done.
    @Override
    public void insertValueInTransactionlessList(String key, Object schema) {
        log.info("Storing schema for key without a transaction {}", key);
        try {
            if (schema == null) {
                redisCommands.rpop(key);
            } else {
                redisCommands.rpush(key, objectMapper.writeValueAsString(schema));
            }
        } catch (JsonProcessingException e) {
            log.error("Jackson object mapping/serialization failed.", e);
            throw new AtlasNonRetryableError(e);
        }
    }

    @Override
    public void updateHeadInList(String key, Object schema) {
        store(key, schema, ListStoreOperation.UPDATE);
    }

    //This is done explicitly due to how rare the use case is. Should generally never be done.
    @Override
    public void updateHeadInTransactionlessList(String key, Object schema) throws AtlasBadKeyError {
        if (redisCommands.exists(key) == 0) {
            throw new AtlasBadKeyError();
        }
        log.info("Storing schema for key without a transaction {}", key);
        try {
            if (schema == null) {
                redisCommands.lpop(key);
            } else {
                redisCommands.lpush(key, objectMapper.writeValueAsString(schema));
            }
        } catch (JsonProcessingException e) {
            log.error("Jackson object mapping/serialization failed.", e);
            throw new AtlasNonRetryableError(e);
        }
    }

    private void store(String key, Object schema, ListStoreOperation listStoreOperation) {
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
            if (result.wasDiscarded()) {
                throw new AtlasRetryableError("The transaction was interrupted");
            }
        } catch (JsonProcessingException e) {
            log.error("Jackson object mapping/serialization failed.", e);
            throw new AtlasNonRetryableError(e);
        }
    }

    @Override
    public PipelineInfo fetchLatestPipelineInfo(String key) {
        return (PipelineInfo) fetch(key, ObjectType.PIPELINE_INFO, -1);
    }

    @Override
    public List<PipelineInfo> fetchPipelineInfoList(String key, int increment) {
        var pipelineInfoList = (List<PipelineInfo>) fetch(key, ObjectType.PIPELINE_INFO_LIST, increment);
        if (pipelineInfoList == null) return new ArrayList<>();
        return pipelineInfoList;
    }

    @Override
    public TeamSchema fetchTeamSchema(String key) {
        return (TeamSchema) fetch(key, ObjectType.TEAM_SCHEMA, -1);
    }

    @Override
    public List<String> fetchStringList(String key) {
        return (List<String>) fetch(key, ObjectType.STRING_LIST, -1);
    }

    @Override
    public ClusterSchema fetchClusterSchema(String key) {
        return (ClusterSchema) fetch(key, ObjectType.CLUSTER_SCHEMA, -1);
    }

    @Override
    public ClusterSchema fetchClusterSchemaTransactionless(String key) {
        return (ClusterSchema) fetchTransactionless(key, ObjectType.CLUSTER_SCHEMA);
    }

    @Override
    public List<Log> fetchLogList(String key, int increment) {
        var logList = (List<Log>) fetch(key, ObjectType.LOG_LIST, increment);
        if (logList == null) return new ArrayList<>();
        return logList;
    }

    @Override
    public Log fetchLatestLog(String key) {
        return (Log) fetch(key, ObjectType.SINGLE_LOG, -1);
    }

    @Override
    public DeploymentLog fetchLatestDeploymentLog(String key) {
        var log = fetchLatestLog(key);
        if (log == null) return null;
        if (log instanceof DeploymentLog) {
            return (DeploymentLog) log;
        } else {
            var idx = 0;
            var logIncrement = 0;
            var deploymentLogList = fetchLogList(key, logIncrement);
            while (idx < deploymentLogList.size()) {
                if (deploymentLogList.get(idx) instanceof DeploymentLog) {
                    return (DeploymentLog) deploymentLogList.get(idx);
                }
                idx++;
                if (idx == deploymentLogList.size()) {
                    logIncrement++;
                    deploymentLogList = fetchLogList(key, logIncrement);
                    idx = 0;
                }
            }
            return null;
        }
    }

    @Override
    public RemediationLog fetchLatestRemediationLog(String key) {
        var log = fetchLatestLog(key);
        if (log instanceof RemediationLog) {
            return (RemediationLog) log;
        } else {
            var idx = 0;
            var logIncrement = 0;
            var logList = fetchLogList(key, logIncrement);
            while (idx < logList.size()) {
                if (logList.get(idx) instanceof RemediationLog) {
                    return (RemediationLog) logList.get(idx);
                }
                idx++;
                if (idx == logList.size()) {
                    logIncrement++;
                    logList = fetchLogList(key, logIncrement);
                    idx = 0;
                }
            }
            return null;
        }
    }

    @Override
    public StepMetadata fetchMetadata(String key) {
        return (StepMetadata) fetch(key, ObjectType.METADATA, -1);
    }

    @Override
    public ClientRequestPacket fetchHeadInClientRequestList(String key) throws AtlasBadKeyError {
        if (redisCommands.exists(key) == 0) {
            return null;
        }
        return (ClientRequestPacket) fetchTransactionless(key, ObjectType.CLIENT_REQUEST);
    }

    private Object fetchTransactionless(String key, ObjectType objectType) {
        try {
            log.info("Fetching schema for key without transaction {}", key);
            var exists = redisCommands.exists(key);
            //If the key doesn't exist, return null (1 is exists, 0 is does not exist)
            if (exists == 0) {
                return null;
            }
            if (objectType == ObjectType.CLUSTER_SCHEMA) {
                var result = redisCommands.get(key);
                return objectMapper.readValue(result, ClusterSchema.class);
            }
            else if (objectType == ObjectType.CLIENT_REQUEST) {
                var result = redisCommands.lindex(key, 0);
                if (result == null) return null;
                return objectMapper.readValue(result, ClientRequestPacket.class);
            }
        } catch (JsonProcessingException e) {
            log.error("Jackson object mapping/serialization failed.", e);
            throw new AtlasNonRetryableError(e);
        }
        throw new AtlasNonRetryableError("None of the ObjectTypes were matched");
    }

    private Object fetch(String key, ObjectType objectType, int increment) {
        try {
            log.info("Fetching schema for key {}", key);
            //This will ensure that the list of watched keys is kept to one, so no leakage will happen.
            redisCommands.unwatch();
            redisCommands.watch(key);
            currentWatchedKey = key;
            var exists = redisCommands.exists(key);
            //If the key doesn't exist, return null (1 is exists, 0 is does not exist)
            if (exists == 0) {
                return null;
            } else if (objectType == ObjectType.TEAM_SCHEMA) {
                var result = redisCommands.get(key);
                return objectMapper.readValue(result, TeamSchema.class);
            } else if (objectType == ObjectType.STRING_LIST) {
                var result = redisCommands.get(key);
                return objectMapper.readValue(result, objectMapper.getTypeFactory().constructCollectionType(List.class, String.class));
            } else if (objectType == ObjectType.LOG_LIST) {
                //TODO: As logs get longer and longer, we cant be fetching a list of 100. We need to find a better way to get chunks of logs as needed.
                var startIdx = increment * LOG_INCREMENT;
                var result = redisCommands.lrange(key, startIdx, startIdx + LOG_INCREMENT - 1);
                var deploymentLogList = new ArrayList<Log>();
                for (var string : result) {
                    var deploymentLog = objectMapper.readValue(string, Log.class);
                    deploymentLogList.add(deploymentLog);
                }
                return deploymentLogList;
            } else if (objectType == ObjectType.PIPELINE_INFO_LIST) {
                //TODO: As logs get longer and longer, we cant be fetching a list of 100. We need to find a better way to get chunks of logs as needed.
                var startIdx = increment * LOG_INCREMENT;
                var result = redisCommands.lrange(key, startIdx, startIdx + LOG_INCREMENT - 1);
                var pipelineInfoList = new ArrayList<PipelineInfo>();
                for (var string : result) {
                    var pipelineInfo = objectMapper.readValue(string, PipelineInfo.class);
                    pipelineInfoList.add(pipelineInfo);
                }
                return pipelineInfoList;
            } else if (objectType == ObjectType.SINGLE_LOG) {
                var result = redisCommands.lindex(key, 0);
                return objectMapper.readValue(result, Log.class);
            } else if (objectType == ObjectType.PIPELINE_INFO) {
                var result = redisCommands.lindex(key, 0);
                return objectMapper.readValue(result, PipelineInfo.class);
            } else if (objectType == ObjectType.CLUSTER_SCHEMA) {
                var result = redisCommands.get(key);
                return objectMapper.readValue(result, ClusterSchema.class);
            } else if (objectType == ObjectType.METADATA) {
                var result = redisCommands.get(key);
                return objectMapper.readValue(result, StepMetadata.class);
            }
        } catch (JsonProcessingException e) {
            log.error("Jackson object mapping/serialization failed.", e);
            throw new AtlasNonRetryableError(e);
        }
        throw new AtlasNonRetryableError("None of the ObjectTypes were matched");
    }
}
