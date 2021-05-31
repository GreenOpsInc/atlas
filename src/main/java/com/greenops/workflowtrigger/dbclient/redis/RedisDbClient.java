package com.greenops.workflowtrigger.dbclient.redis;

import com.greenops.workflowtrigger.api.model.git.GitRepoSchema;
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

    public RedisDbClient() {
        var host = System.getenv(REDIS_HOST_VARIABLE);
        var port = System.getenv(REDIS_PORT_VARIABLE);
        client = new RedisClient(RedisURI.create("redis://" + host + ":" + port)); //Pattern is redis://password@host:port
    }

    @Override
    public boolean store(GitRepoSchema gitRepoSchema) {
        var connection = client.connect();
        //TODO: The key is wrong. It should be updated when the pipeline schema POJO is written
        var result = connection.set(gitRepoSchema.getGitRepo(), gitRepoSchema.convertToJson().toString());
        connection.close();
        return result.equals(REDIS_SUCCESS_MESSAGE);
    }
}
