package com.greenops.workfloworchestrator.config;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.dataformat.yaml.YAMLFactory;
import com.greenops.workfloworchestrator.datamodel.event.ClientCompletionEvent;
import com.greenops.workfloworchestrator.datamodel.git.GitCredMachineUser;
import com.greenops.workfloworchestrator.datamodel.git.GitCredOpen;
import com.greenops.workfloworchestrator.datamodel.git.GitCredToken;
import com.greenops.workfloworchestrator.datamodel.git.GitRepoSchema;
import com.greenops.workfloworchestrator.datamodel.mixin.event.ClientCompletionEventMixin;
import com.greenops.workfloworchestrator.datamodel.mixin.git.GitCredMachineUserMixin;
import com.greenops.workfloworchestrator.datamodel.mixin.git.GitCredTokenMixin;
import com.greenops.workfloworchestrator.datamodel.mixin.git.GitRepoSchemaMixin;
import com.greenops.workfloworchestrator.datamodel.mixin.pipelinedata.PipelineDataMixin;
import com.greenops.workfloworchestrator.datamodel.mixin.pipelinedata.StepDataMixin;
import com.greenops.workfloworchestrator.datamodel.mixin.pipelinedata.TestMixin;
import com.greenops.workfloworchestrator.datamodel.mixin.pipelineschema.PipelineSchemaMixin;
import com.greenops.workfloworchestrator.datamodel.mixin.pipelineschema.TeamSchemaMixin;
import com.greenops.workfloworchestrator.datamodel.mixin.requests.DeployResponseMixin;
import com.greenops.workfloworchestrator.datamodel.mixin.requests.GetFileRequestMixin;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.*;
import com.greenops.workfloworchestrator.datamodel.pipelineschema.PipelineSchemaImpl;
import com.greenops.workfloworchestrator.datamodel.pipelineschema.TeamSchemaImpl;
import com.greenops.workfloworchestrator.datamodel.requests.DeployResponse;
import com.greenops.workfloworchestrator.datamodel.requests.GetFileRequest;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class SpringConfiguration {

    @Bean
    @Qualifier("yamlObjectMapper")
    ObjectMapper yamlObjectMapper() {
        return new ObjectMapper(new YAMLFactory());
    }

    @Bean
    @Qualifier("eventAndRequestObjectMapper")
    ObjectMapper eventAndRequestObjectMapper() {
        return new ObjectMapper()
                .addMixIn(ClientCompletionEvent.class, ClientCompletionEventMixin.class)
                .addMixIn(GetFileRequest.class, GetFileRequestMixin.class)
                .addMixIn(DeployResponse.class, DeployResponseMixin.class);
    }

    @Bean
    @Qualifier("objectMapper")
    ObjectMapper objectMapper() {
        return new ObjectMapper()
                .addMixIn(PipelineDataImpl.class, PipelineDataMixin.class)
                .addMixIn(StepDataImpl.class, StepDataMixin.class)
                .addMixIn(CustomTest.class, TestMixin.class)
                .addMixIn(PipelineSchemaImpl.class, PipelineSchemaMixin.class)
                .addMixIn(TeamSchemaImpl.class, TeamSchemaMixin.class)
                .addMixIn(GitRepoSchema.class, GitRepoSchemaMixin.class)
                .addMixIn(GitCredMachineUser.class, GitCredMachineUserMixin.class)
                .addMixIn(GitCredToken.class, GitCredTokenMixin.class)
                .addMixIn(GitCredOpen.class, GitCredOpen.class);
    }
}
