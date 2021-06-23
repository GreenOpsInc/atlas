package com.greenops.workfloworchestrator.config;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.dataformat.yaml.YAMLFactory;
import com.greenops.workfloworchestrator.datamodel.mixin.pipelinedata.PipelineDataMixin;
import com.greenops.workfloworchestrator.datamodel.mixin.pipelinedata.StepDataMixin;
import com.greenops.workfloworchestrator.datamodel.mixin.pipelinedata.TestMixin;
import com.greenops.workfloworchestrator.datamodel.mixin.requests.DeployResponseMixin;
import com.greenops.workfloworchestrator.datamodel.mixin.requests.GetFileRequestMixin;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.*;
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
    @Qualifier("requestObjectMapper")
    ObjectMapper requestObjectMapper() {
        return new ObjectMapper()
                .addMixIn(GetFileRequest.class, GetFileRequestMixin.class)
                .addMixIn(DeployResponse.class, DeployResponseMixin.class);
    }

    @Bean
    @Qualifier("objectMapper")
    ObjectMapper objectMapper() {
        return new ObjectMapper()
                .addMixIn(PipelineDataImpl.class, PipelineDataMixin.class)
                .addMixIn(StepDataImpl.class, StepDataMixin.class)
                .addMixIn(CustomTest.class, TestMixin.class);
    }
}
