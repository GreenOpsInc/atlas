package com.greenops.workfloworchestrator.config;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonSerializable;
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
import com.greenops.util.kubernetesclient.KubernetesClient;
import com.greenops.util.kubernetesclient.KubernetesClientImpl;
import com.greenops.util.tslmanager.TLSManager;
import com.greenops.util.tslmanager.TLSManagerImpl;
import com.greenops.workfloworchestrator.datamodel.mixin.pipelinedata.*;
import com.greenops.workfloworchestrator.datamodel.mixin.requests.*;
import com.greenops.workfloworchestrator.datamodel.pipelinedata.*;
import com.greenops.workfloworchestrator.datamodel.requests.*;
import com.greenops.workfloworchestrator.ingest.kafka.KafkaClient;
import lombok.extern.slf4j.Slf4j;
import org.apache.kafka.clients.CommonClientConfigs;
import org.apache.kafka.clients.consumer.ConsumerConfig;
import org.apache.kafka.common.config.SslConfigs;
import org.apache.kafka.common.serialization.StringDeserializer;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.kafka.config.ConcurrentKafkaListenerContainerFactory;
import org.springframework.kafka.core.*;
import org.springframework.kafka.listener.ContainerAwareErrorHandler;
import org.springframework.kafka.listener.SeekToCurrentErrorHandler;
import org.springframework.kafka.support.serializer.JsonDeserializer;
import org.springframework.scheduling.config.Task;
import org.springframework.util.backoff.FixedBackOff;

import java.io.IOException;
import java.util.HashMap;
import java.util.Map;

@Slf4j
@Configuration
public class SpringConfiguration {

    private static final String KAFKA_SSL_TRUSTSTORE_LOCATION_CONFIG = System.getenv("KAFKA_SSL_TRUSTSTORE_LOCATION_CONFIG");
    private static final String KAFKA_SSL_KEYSTORE_LOCATION_CONFIG = System.getenv("KAFKA_SSL_KEYSTORE_LOCATION_CONFIG");

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

    @Bean
    public ProducerFactory<String, String> producerFactory(
            String bootstrapServers
    ) {

        Map<String, Object> configProps = new HashMap<>();

        // TODO: on app start try to get cert and key from secrets
        //      1. get cert & key from secrets
        //      2. if secret not found return default producer without ssl enabled
        //      3. if secret found create keystore and save cert in it
        //      4. keystore location values should be stored in the env vars
        //      5. get those env vars and update producer config
        //      6. add watcher for secret
        //      7. on app start if secret is not available but keystore exists delete the keystore
        //      8. on app start event if keystore and secret exists update the keystore
        //      9. on secret change halt the application and new config should be generated on app start

        KubernetesClient kclient;
        try {
            // TODO: somehow get object mapper from beans
            ObjectMapper objectMapper = new ObjectMapper();
            kclient = new KubernetesClientImpl(objectMapper);
        } catch (IOException exc) {
            // TODO: log and stop server
            throw new RuntimeException("Could not initialize Kubernetes Client", exc);
        }

        boolean keystoreExists;
        try {
            TLSManager tlsManager = new TLSManagerImpl(kclient, null, null);
            keystoreExists = tlsManager.updateKafkaKeystore(KAFKA_SSL_TRUSTSTORE_LOCATION_CONFIG, KAFKA_SSL_KEYSTORE_LOCATION_CONFIG);
            if (keystoreExists) {
                tlsManager.watchKafkaKeystore(KAFKA_SSL_TRUSTSTORE_LOCATION_CONFIG,KAFKA_SSL_KEYSTORE_LOCATION_CONFIG);
            }
        } catch (Exception exc) {
            // TODO: log and stop server
            throw new RuntimeException("Could not configure Kafka with TLS configuration", exc);
        }

        if (keystoreExists) {
            configProps.put(CommonClientConfigs.SECURITY_PROTOCOL_CONFIG, "SSL");
            configProps.put(SslConfigs.SSL_TRUSTSTORE_LOCATION_CONFIG, KAFKA_SSL_TRUSTSTORE_LOCATION_CONFIG);
            configProps.put(SslConfigs.SSL_KEYSTORE_LOCATION_CONFIG, KAFKA_SSL_KEYSTORE_LOCATION_CONFIG);
        }
        return new DefaultKafkaProducerFactory<>(configProps);
    }

    @Bean
    public ConsumerFactory<String, String> consumerFactory() {
        Map<String, Object> configProps = new HashMap<>();

        // TODO: on app start try to get cert and key from secrets
        //      1. get cert & key from secrets
        //      2. if secret not found return default producer without ssl enabled
        //      3. if secret found create keystore and save cert in it
        //      4. keystore location values should be stored in the env vars
        //      5. get those env vars and update producer config
        //      6. add watcher for secret
        //      7. on app start if secret is not available but keystore exists delete the keystore
        //      8. on app start event if keystore and secret exists update the keystore
        //      9. on secret change halt the application and new config should be generated on app start

        KubernetesClient kclient;
        try {
            // TODO: somehow get object mapper from beans
            ObjectMapper objectMapper = new ObjectMapper();
            kclient = new KubernetesClientImpl(objectMapper);
        } catch (IOException exc) {
            // TODO: log and stop server
            throw new RuntimeException("Could not initialize Kubernetes Client", exc);
        }

        boolean keystoreExists;
        try {
            TLSManager tlsManager = new TLSManagerImpl(kclient, null, null);
            keystoreExists = tlsManager.updateKafkaKeystore(KAFKA_SSL_TRUSTSTORE_LOCATION_CONFIG, KAFKA_SSL_KEYSTORE_LOCATION_CONFIG);
        } catch (Exception exc) {
            // TODO: log and stop server
            throw new RuntimeException("Could not configure Kafka with TLS configuration", exc);
        }

        if (keystoreExists) {
            configProps.put(CommonClientConfigs.SECURITY_PROTOCOL_CONFIG, "SSL");
            configProps.put(SslConfigs.SSL_TRUSTSTORE_LOCATION_CONFIG, KAFKA_SSL_TRUSTSTORE_LOCATION_CONFIG);
            configProps.put(SslConfigs.SSL_KEYSTORE_LOCATION_CONFIG, KAFKA_SSL_KEYSTORE_LOCATION_CONFIG);
        }
        return new DefaultKafkaConsumerFactory<>(configProps);
    }

    @Bean
    public ConcurrentKafkaListenerContainerFactory<String, String> kafkaListenerContainerFactory() {
        ConcurrentKafkaListenerContainerFactory<String, String> factory =
                new ConcurrentKafkaListenerContainerFactory<>();
        factory.setConsumerFactory(consumerFactory());
        return factory;
    }

    @Bean
    public KafkaTemplate<String, String> kafkaTemplate(
            @Value("${spring.kafka.producer.bootstrap-servers}") String bootstrapServers,
            ) {
        return new KafkaTemplate<>(producerFactory(bootstrapServers));
    }

    // TODO: add all values from application.yml
    private Map<String, Object> getKafkaConfigProps() {
        Map<String, Object> configProps = new HashMap<>();

        // TODO: on app start try to get cert and key from secrets
        //      1. get cert & key from secrets
        //      2. if secret not found return default producer without ssl enabled
        //      3. if secret found create keystore and save cert in it
        //      4. keystore location values should be stored in the env vars
        //      5. get those env vars and update producer config
        //      6. add watcher for secret
        //      7. on app start if secret is not available but keystore exists delete the keystore
        //      8. on app start event if keystore and secret exists update the keystore
        //      9. on secret change halt the application and new config should be generated on app start

        KubernetesClient kclient;
        try {
            // TODO: somehow get object mapper from beans
            ObjectMapper objectMapper = new ObjectMapper();
            kclient = new KubernetesClientImpl(objectMapper);
        } catch (IOException exc) {
            // TODO: log and stop server
            throw new RuntimeException("Could not initialize Kubernetes Client", exc);
        }

        boolean keystoreExists;
        try {
            TLSManager tlsManager = new TLSManagerImpl(kclient, null, null);
            keystoreExists = tlsManager.updateKafkaKeystore(KAFKA_SSL_TRUSTSTORE_LOCATION_CONFIG, KAFKA_SSL_KEYSTORE_LOCATION_CONFIG);
            if (keystoreExists) {
                tlsManager.watchKafkaKeystore(KAFKA_SSL_TRUSTSTORE_LOCATION_CONFIG,KAFKA_SSL_KEYSTORE_LOCATION_CONFIG);
            }
        } catch (Exception exc) {
            // TODO: log and stop server
            throw new RuntimeException("Could not configure Kafka with TLS configuration", exc);
        }

        if (keystoreExists) {
            configProps.put(CommonClientConfigs.SECURITY_PROTOCOL_CONFIG, "SSL");
            configProps.put(SslConfigs.SSL_TRUSTSTORE_LOCATION_CONFIG, KAFKA_SSL_TRUSTSTORE_LOCATION_CONFIG);
            configProps.put(SslConfigs.SSL_KEYSTORE_LOCATION_CONFIG, KAFKA_SSL_KEYSTORE_LOCATION_CONFIG);
        }
        return configProps;
    }
}
