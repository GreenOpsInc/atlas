package com.greenops.workfloworchestrator.config;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.dataformat.yaml.YAMLFactory;
import com.greenops.util.datamodel.auditlog.DeploymentLog;
import com.greenops.util.datamodel.auditlog.PipelineInfo;
import com.greenops.util.datamodel.auditlog.RemediationLog;
import com.greenops.util.datamodel.clientmessages.*;
import com.greenops.util.datamodel.event.*;
import com.greenops.util.datamodel.git.*;
import com.greenops.util.datamodel.mixin.auditlog.PipelineInfoMixin;
import com.greenops.util.datamodel.request.DeployResponse;
import com.greenops.util.datamodel.metadata.StepMetadata;
import com.greenops.util.datamodel.mixin.auditlog.DeploymentLogMixin;
import com.greenops.util.datamodel.mixin.auditlog.RemediationLogMixin;
import com.greenops.util.datamodel.mixin.clientmessages.*;
import com.greenops.util.datamodel.mixin.event.*;
import com.greenops.util.datamodel.mixin.git.*;
import com.greenops.util.datamodel.mixin.metadata.StepMetadataMixin;
import com.greenops.util.datamodel.mixin.pipeline.PipelineSchemaMixin;
import com.greenops.util.datamodel.mixin.pipeline.TeamSchemaMixin;
import com.greenops.util.datamodel.mixin.request.*;
import com.greenops.util.datamodel.pipeline.PipelineSchemaImpl;
import com.greenops.util.datamodel.pipeline.TeamSchemaImpl;
import com.greenops.util.datamodel.request.*;
import com.greenops.util.dbclient.DbClient;
import com.greenops.util.dbclient.redis.RedisDbClient;
import com.greenops.util.error.AtlasNonRetryableError;
import com.greenops.util.error.AtlasRetryableError;
import com.greenops.workfloworchestrator.datamodel.mixin.pipelinedata.*;
import com.greenops.workfloworchestrator.datamodel.mixin.requests.*;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.*;
import com.greenops.workfloworchestrator.datamodel.requests.*;
import com.greenops.workfloworchestrator.ingest.kafka.KafkaClient;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.kafka.listener.ContainerAwareErrorHandler;
import org.springframework.kafka.listener.SeekToCurrentErrorHandler;
import org.springframework.util.backoff.FixedBackOff;

@Slf4j
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
                .addMixIn(ResourceStatus.class, ResourceStatusMixin.class)
                .addMixIn(TestCompletionEvent.class, TestCompletionEventMixin.class)
                .addMixIn(FailureEvent.class, FailureEventMixin.class)
                .addMixIn(ApplicationInfraTriggerEvent.class, ApplicationInfraTriggerEventMixin.class)
                .addMixIn(ApplicationInfraCompletionEvent.class, ApplicationInfraCompletionEventMixin.class)
                .addMixIn(ClientRequestPacket.class, ClientRequestPacketMixin.class)
                .addMixIn(ClientDeleteByConfigRequest.class, ClientDeleteByConfigRequestMixin.class)
                .addMixIn(ClientDeleteByGvkRequest.class, ClientDeleteByGvkRequestMixin.class)
                .addMixIn(ClientDeployAndWatchRequest.class, ClientDeployAndWatchRequestMixin.class)
                .addMixIn(ClientDeployNamedArgoAppAndWatchRequest.class, ClientDeployNamedArgoAppAndWatchRequestMixin.class)
                .addMixIn(ClientDeployNamedArgoApplicationRequest.class, ClientDeployNamedArgoApplicationRequestMixin.class)
                .addMixIn(ClientDeployRequest.class, ClientDeployRequestMixin.class)
                .addMixIn(ClientRollbackAndWatchRequest.class, ClientRollbackAndWatchRequestMixin.class)
                .addMixIn(GetFileRequest.class, GetFileRequestMixin.class)
                .addMixIn(WatchRequest.class, WatchRequestMixin.class)
                .addMixIn(KubernetesCreationRequest.class, KubernetesCreationRequestMixin.class)
                .addMixIn(ResourcesGvkRequest.class, ResourcesGvkRequestMixin.class)
                .addMixIn(ResourceGvk.class, ResourceGvkMixin.class)
                .addMixIn(DeployResponse.class, DeployResponseMixin.class)
                .addMixIn(TriggerStepEvent.class, TriggerStepEventMixin.class)
                .addMixIn(PipelineTriggerEvent.class, PipelineTriggerEventMixin.class);
    }

    @Bean
    @Qualifier("objectMapper")
    ObjectMapper objectMapper() {
        return new ObjectMapper()
                .addMixIn(PipelineDataImpl.class, PipelineDataMixin.class)
                .addMixIn(StepDataImpl.class, StepDataMixin.class)
                .addMixIn(InjectScriptTest.class, InjectScriptTestMixin.class)
                .addMixIn(CustomJobTest.class, CustomJobTestMixin.class)
                .addMixIn(ArgoWorkflowTask.class, ArgoWorkflowTaskMixin.class)
                .addMixIn(PipelineSchemaImpl.class, PipelineSchemaMixin.class)
                .addMixIn(TeamSchemaImpl.class, TeamSchemaMixin.class)
                .addMixIn(GitRepoSchema.class, GitRepoSchemaMixin.class)
                .addMixIn(GitCredMachineUser.class, GitCredMachineUserMixin.class)
                .addMixIn(GitCredToken.class, GitCredTokenMixin.class)
                .addMixIn(GitCredOpen.class, GitCredOpenMixin.class)
                .addMixIn(DeploymentLog.class, DeploymentLogMixin.class)
                .addMixIn(RemediationLog.class, RemediationLogMixin.class)
                .addMixIn(ResourceStatus.class, ResourceStatusMixin.class)
                .addMixIn(StepMetadata.class, StepMetadataMixin.class)
                .addMixIn(PipelineInfo.class, PipelineInfoMixin.class)
                .addMixIn(ArgoRepoSchema.class, ArgoRepoSchemaMixin.class)
                .addMixIn(ClientDeleteByConfigRequest.class, ClientDeleteByConfigRequestMixin.class)
                .addMixIn(ClientDeleteByGvkRequest.class, ClientDeleteByGvkRequestMixin.class)
                .addMixIn(ClientDeployAndWatchRequest.class, ClientDeployAndWatchRequestMixin.class)
                .addMixIn(ClientDeployRequest.class, ClientDeployRequestMixin.class)
                .addMixIn(ClientRollbackAndWatchRequest.class, ClientRollbackAndWatchRequestMixin.class)
                .addMixIn(ClientSelectiveSyncAndWatchRequest.class, ClientSelectiveSyncAndWatchRequestMixin.class)
                .addMixIn(ResourcesGvkRequest.class, ResourcesGvkRequestMixin.class)
                .addMixIn(ResourceGvk.class, ResourceGvkMixin.class);
    }

    @Bean
    DbClient dbClient(@Value("${application.redis-url}") String redisUrl, ObjectMapper objectMapper) {
        return new RedisDbClient(redisUrl, objectMapper);
    }

    @Bean
    ContainerAwareErrorHandler errorHandler(KafkaClient kafkaClient) {
        var errorHandler =
                new SeekToCurrentErrorHandler((record, exception) -> {
                    //send to DLQ
                    log.info(exception.getMessage(), exception.getCause());
                    kafkaClient.sendMessageToDlq((String) record.value());
                    try {
                        var event = eventAndRequestObjectMapper().readValue((String) record.value(), Event.class);
                        //If its a failure event, chances are the error will keep looping forever
                        if (!(event instanceof FailureEvent)) {
                            var failureEvent = new FailureEvent(
                                    event.getOrgName(),
                                    event.getTeamName(),
                                    event.getPipelineName(),
                                    event.getPipelineUvn(),
                                    event.getStepName(),
                                    null,
                                    event.getClass().getName(),
                                    exception.getMessage()
                            );
                            kafkaClient.sendMessage(failureEvent);
                        }
                    } catch (JsonProcessingException e) {
                        log.info("Couldn't deserialize event to send failure event", e.getCause());
                    }
                }, new FixedBackOff(100L, 5L));
        errorHandler.addNotRetryableException(AtlasNonRetryableError.class);
        return errorHandler;
    }
}
