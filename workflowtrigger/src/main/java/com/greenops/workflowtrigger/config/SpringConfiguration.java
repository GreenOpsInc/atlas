package com.greenops.workflowtrigger.config;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.dataformat.yaml.YAMLFactory;
import com.greenops.util.datamodel.auditlog.DeploymentLog;
import com.greenops.util.datamodel.auditlog.RemediationLog;
import com.greenops.util.datamodel.cluster.ClusterSchema;
import com.greenops.util.datamodel.event.ClientCompletionEvent;
import com.greenops.util.datamodel.git.GitCredMachineUser;
import com.greenops.util.datamodel.git.GitCredToken;
import com.greenops.util.datamodel.git.GitRepoSchema;
import com.greenops.util.datamodel.mixin.auditlog.DeploymentLogMixin;
import com.greenops.util.datamodel.mixin.auditlog.RemediationLogMixin;
import com.greenops.util.datamodel.mixin.cluster.ClusterSchemaMixin;
import com.greenops.util.datamodel.mixin.event.ClientCompletionEventMixin;
import com.greenops.util.datamodel.mixin.git.GitCredMachineUserMixin;
import com.greenops.util.datamodel.mixin.git.GitCredTokenMixin;
import com.greenops.util.datamodel.mixin.git.GitRepoSchemaMixin;
import com.greenops.util.datamodel.mixin.pipeline.PipelineSchemaMixin;
import com.greenops.util.datamodel.mixin.pipeline.TeamSchemaMixin;
import com.greenops.util.datamodel.mixin.pipelinestatus.FailedStepMixin;
import com.greenops.util.datamodel.mixin.pipelinestatus.PipelineStatusMixin;
import com.greenops.util.datamodel.pipeline.PipelineSchemaImpl;
import com.greenops.util.datamodel.pipeline.TeamSchemaImpl;
import com.greenops.util.datamodel.pipelinestatus.FailedStep;
import com.greenops.util.datamodel.pipelinestatus.PipelineStatus;
import com.greenops.util.dbclient.DbClient;
import com.greenops.util.dbclient.redis.RedisDbClient;
import com.greenops.workflowtrigger.datamodel.mixin.pipelinedata.PipelineDataMixin;
import com.greenops.workflowtrigger.datamodel.mixin.pipelinedata.StepDataMixin;
import com.greenops.workflowtrigger.datamodel.pipelinedata.PipelineDataImpl;
import com.greenops.workflowtrigger.datamodel.pipelinedata.StepDataImpl;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.kafka.core.ProducerFactory;

@Configuration
public class SpringConfiguration {

    @Bean
    @Qualifier("yamlObjectMapper")
    ObjectMapper yamlObjectMapper() {
        return new ObjectMapper(new YAMLFactory());
    }

    @Bean
    public ObjectMapper objectMapper() {
        return new ObjectMapper()
                .addMixIn(TeamSchemaImpl.class, TeamSchemaMixin.class)
                .addMixIn(PipelineSchemaImpl.class, PipelineSchemaMixin.class)
                .addMixIn(GitRepoSchema.class, GitRepoSchemaMixin.class)
                .addMixIn(GitCredMachineUser.class, GitCredMachineUserMixin.class)
                .addMixIn(GitCredToken.class, GitCredTokenMixin.class)
                .addMixIn(ClientCompletionEvent.class, ClientCompletionEventMixin.class)
                .addMixIn(DeploymentLog.class, DeploymentLogMixin.class)
                .addMixIn(RemediationLog.class, RemediationLogMixin.class)
                .addMixIn(ClusterSchema.class, ClusterSchemaMixin.class)
                .addMixIn(FailedStep.class, FailedStepMixin.class)
                .addMixIn(PipelineStatus.class, PipelineStatusMixin.class)
                .addMixIn(PipelineDataImpl.class, PipelineDataMixin.class)
                .addMixIn(StepDataImpl.class, StepDataMixin.class);

    }

    @Bean
    public KafkaTemplate<String, String> kafkaTemplate(ProducerFactory<String, String> producerFactory) {
        return new KafkaTemplate<>(producerFactory);
    }

    @Bean
    public DbClient dbClient(@Value("${application.redis-url}") String redisUrl, ObjectMapper objectMapper) {
        return new RedisDbClient(redisUrl, objectMapper);
    }
}

