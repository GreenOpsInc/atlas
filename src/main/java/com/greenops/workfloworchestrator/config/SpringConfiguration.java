package com.greenops.workfloworchestrator.config;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.dataformat.yaml.YAMLFactory;
import com.greenops.workfloworchestrator.datamodel.mixin.PipelineDataMixin;
import com.greenops.workfloworchestrator.datamodel.mixin.StepDataMixin;
import com.greenops.workfloworchestrator.datamodel.mixin.TestMixin;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.*;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class SpringConfiguration {

    @Bean
    ObjectMapper yamlObjectMapper() {
        return new ObjectMapper(new YAMLFactory());
    }

    @Bean
    ObjectMapper objectMapper() {
        return new ObjectMapper()
                .addMixIn(PipelineDataImpl.class, PipelineDataMixin.class)
                .addMixIn(StepDataImpl.class, StepDataMixin.class)
                .addMixIn(CustomTest.class, TestMixin.class);
    }
}
