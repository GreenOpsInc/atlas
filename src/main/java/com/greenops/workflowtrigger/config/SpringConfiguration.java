package com.greenops.workflowtrigger.config;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.workflowtrigger.api.model.auditlog.DeploymentLog;
import com.greenops.workflowtrigger.api.model.cluster.ClusterSchema;
import com.greenops.workflowtrigger.api.model.event.ClientCompletionEvent;
import com.greenops.workflowtrigger.api.model.git.GitCredMachineUser;
import com.greenops.workflowtrigger.api.model.git.GitCredToken;
import com.greenops.workflowtrigger.api.model.git.GitRepoSchema;
import com.greenops.workflowtrigger.api.model.mixin.auditlog.DeploymentLogMixin;
import com.greenops.workflowtrigger.api.model.mixin.cluster.ClusterSchemaMixin;
import com.greenops.workflowtrigger.api.model.mixin.event.ClientCompletionEventMixin;
import com.greenops.workflowtrigger.api.model.mixin.git.GitCredMachineUserMixin;
import com.greenops.workflowtrigger.api.model.mixin.git.GitCredTokenMixin;
import com.greenops.workflowtrigger.api.model.mixin.git.GitRepoSchemaMixin;
import com.greenops.workflowtrigger.api.model.mixin.pipeline.PipelineSchemaMixin;
import com.greenops.workflowtrigger.api.model.mixin.pipeline.TeamSchemaMixin;
import com.greenops.workflowtrigger.api.model.pipeline.PipelineSchemaImpl;
import com.greenops.workflowtrigger.api.model.pipeline.TeamSchemaImpl;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.kafka.core.ProducerFactory;

@Configuration
public class SpringConfiguration {

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
                .addMixIn(ClusterSchema.class, ClusterSchemaMixin.class);

    }

    @Bean
    public KafkaTemplate<String, String> kafkaTemplate(ProducerFactory<String, String> producerFactory) {
        return new KafkaTemplate<>(producerFactory);
    }
}

