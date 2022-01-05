package com.greenops.pipelinereposerver.config;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.util.datamodel.git.GitCredMachineUser;
import com.greenops.util.datamodel.git.GitCredToken;
import com.greenops.util.datamodel.git.GitRepoSchema;
import com.greenops.util.datamodel.mixin.git.GitCredMachineUserMixin;
import com.greenops.util.datamodel.mixin.git.GitCredTokenMixin;
import com.greenops.util.datamodel.mixin.git.GitRepoSchemaMixin;
import com.greenops.util.datamodel.mixin.pipeline.PipelineSchemaMixin;
import com.greenops.util.datamodel.mixin.pipeline.TeamSchemaMixin;
import com.greenops.util.datamodel.mixin.request.GetFileRequestMixin;
import com.greenops.util.datamodel.pipeline.PipelineSchemaImpl;
import com.greenops.util.datamodel.pipeline.TeamSchemaImpl;
import com.greenops.util.datamodel.request.GetFileRequest;
import com.greenops.util.dbclient.DbClient;
import com.greenops.util.dbclient.redis.RedisDbClient;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

// TODO: server configuration
//      1. add bean to:
//          a. read config from secrets
//          b. if not available create self signed
//          c. create keystore
//          d. use keystore for tomcat
//      2. create watcher to watch kubernetes secret updates
//          a. on create/update/delete reload the server
//          b. after server is reloaded it will pull new config from secrets and create the server
//          c. the same could be done for kafka (I guess)

@Configuration
public class SpringConfiguration {

    @Bean
    ObjectMapper objectMapper() {
        return new ObjectMapper()
                .addMixIn(TeamSchemaImpl.class, TeamSchemaMixin.class)
                .addMixIn(PipelineSchemaImpl.class, PipelineSchemaMixin.class)
                .addMixIn(GitRepoSchema.class, GitRepoSchemaMixin.class)
                .addMixIn(GitCredMachineUser.class, GitCredMachineUserMixin.class)
                .addMixIn(GitCredToken.class, GitCredTokenMixin.class)
                .addMixIn(GetFileRequest.class, GetFileRequestMixin.class);
    }

    @Bean
    DbClient dbClient(@Value("${application.redis-url}") String redisUrl, ObjectMapper objectMapper) {
        return new RedisDbClient(redisUrl, objectMapper);
    }
}
