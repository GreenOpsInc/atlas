package com.greenops.pipelinereposerver.config;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.pipelinereposerver.api.model.git.GitCredMachineUser;
import com.greenops.pipelinereposerver.api.model.git.GitCredToken;
import com.greenops.pipelinereposerver.api.model.git.GitRepoSchema;
import com.greenops.pipelinereposerver.api.model.mixin.git.GitCredMachineUserMixin;
import com.greenops.pipelinereposerver.api.model.mixin.git.GitCredTokenMixin;
import com.greenops.pipelinereposerver.api.model.mixin.git.GitRepoSchemaMixin;
import com.greenops.pipelinereposerver.api.model.mixin.pipeline.PipelineSchemaMixin;
import com.greenops.pipelinereposerver.api.model.mixin.pipeline.TeamSchemaMixin;
import com.greenops.pipelinereposerver.api.model.mixin.request.GetFileRequestMixin;
import com.greenops.pipelinereposerver.api.model.pipeline.PipelineSchemaImpl;
import com.greenops.pipelinereposerver.api.model.pipeline.TeamSchemaImpl;
import com.greenops.pipelinereposerver.api.model.request.GetFileRequest;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

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
}
