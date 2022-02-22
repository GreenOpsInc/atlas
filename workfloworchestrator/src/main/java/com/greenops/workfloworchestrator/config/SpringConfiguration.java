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
import com.greenops.util.datamodel.mixin.pipelinedata.*;
import com.greenops.util.datamodel.pipelinedata.*;
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
import com.greenops.util.kubernetesclient.KubernetesClient;
import com.greenops.util.kubernetesclient.KubernetesClientImpl;
import com.greenops.util.tslmanager.TLSManager;
import com.greenops.util.tslmanager.TLSManagerImpl;
import com.greenops.workfloworchestrator.ingest.kafka.KafkaClient;
import lombok.extern.slf4j.Slf4j;
import org.apache.kafka.clients.consumer.ConsumerConfig;
import org.apache.kafka.clients.producer.ProducerConfig;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.kafka.config.ConcurrentKafkaListenerContainerFactory;
import org.springframework.kafka.core.*;
import org.springframework.kafka.listener.ContainerAwareErrorHandler;
import org.springframework.kafka.listener.ContainerProperties;
import org.springframework.kafka.listener.SeekToCurrentErrorHandler;
import org.springframework.util.backoff.FixedBackOff;

import java.io.IOException;
import java.util.HashMap;
import java.util.Map;

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
                .addMixIn(GitRepoSchemaInfo.class, GitRepoSchemaInfoMixin.class)
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
    KubernetesClient kubernetesClient(ObjectMapper objectMapper) {
        KubernetesClient kclient;
        try {
            kclient = new KubernetesClientImpl(objectMapper);
        } catch (IOException exc) {
            throw new RuntimeException("Could not initialize Kubernetes Client", exc);
        }
        return kclient;
    }

    @Bean
    TLSManager tlsManager(KubernetesClient kclient) {
        return new TLSManagerImpl(kclient);
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
        errorHandler.addNotRetryableExceptions(AtlasNonRetryableError.class);
        return errorHandler;
    }

    @Bean
    public KafkaTemplate<String, String> kafkaTemplate(
            TLSManager tlsManager,
            @Value("${application.kafka.producer.bootstrap-servers}") String bootstrapServers,
            @Value("${application.kafka.producer.key-serializer}") String keySerializer,
            @Value("${application.kafka.producer.value-serializer}") String valueSerializer,
            @Value("${application.kafka.ssl.keystore-location}") String keystoreLocation,
            @Value("${application.kafka.ssl.truststore-location}") String truststoreLocation
    ) {
        ProducerFactory<String, String> factory = producerFactory(tlsManager,bootstrapServers, keySerializer, valueSerializer, keystoreLocation, truststoreLocation);
        return new KafkaTemplate<>(factory);
    }

    @Bean
    public ConcurrentKafkaListenerContainerFactory<String, String> kafkaListenerContainerFactory(
            TLSManager tlsManager,
            KafkaClient kafkaClient,
            @Value("${application.kafka.consumer.group-id}") String groupId,
            @Value("${application.kafka.consumer.auto-offset-reset}") String autoOffsetReset,
            @Value("${application.kafka.consumer.enable-auto-commit}") String enableAutoCommit,
            @Value("${application.kafka.consumer.bootstrap-servers}") String bootstrapServers,
            @Value("${application.kafka.consumer.key-deserializer}") String keyDeserializer,
            @Value("${application.kafka.consumer.value-deserializer}") String valueDeserializer,
            @Value("${application.kafka.ssl.keystore-location}") String keystoreLocation,
            @Value("${application.kafka.ssl.truststore-location}") String truststoreLocation
    ) {
        ConcurrentKafkaListenerContainerFactory<String, String> factory =
                new ConcurrentKafkaListenerContainerFactory<>();
        ConsumerFactory<String, String> consumerFactory = consumerFactory(tlsManager,groupId, autoOffsetReset, enableAutoCommit, bootstrapServers, keyDeserializer, valueDeserializer, keystoreLocation, truststoreLocation);
        factory.setConsumerFactory(consumerFactory);
        factory.setErrorHandler(errorHandler(kafkaClient));
        factory.getContainerProperties().setAckMode(ContainerProperties.AckMode.MANUAL_IMMEDIATE);
        return factory;
    }

    @Bean
    public ProducerFactory<String, String> producerFactory(TLSManager tlsManager,String bootstrapServers, String keySerializer, String valueSerializer, String keystoreLocation, String truststoreLocation) {
        return new DefaultKafkaProducerFactory<>(getKafkaProducerConfigProps(tlsManager,bootstrapServers, keySerializer, valueSerializer, keystoreLocation, truststoreLocation));
    }

    @Bean
    public ConsumerFactory<String, String> consumerFactory(TLSManager tlsManager,String groupId, String autoOffsetReset, String enableAutoCommit, String bootstrapServers, String keyDeserializer, String valueDeserializer, String keystoreLocation, String truststoreLocation) {
        return new DefaultKafkaConsumerFactory<>(getKafkaConsumerConfigProps(tlsManager,groupId, autoOffsetReset, enableAutoCommit, bootstrapServers, keyDeserializer, valueDeserializer, keystoreLocation, truststoreLocation));
    }

    private Map<String, Object> getKafkaProducerConfigProps(TLSManager tlsManager,String bootstrapServers, String keySerializer, String valueSerializer, String keystoreLocation, String truststoreLocation) {
        Map<String, Object> configProps = new HashMap<>();
        configProps.put(ProducerConfig.BOOTSTRAP_SERVERS_CONFIG, bootstrapServers);
        configProps.put(ProducerConfig.KEY_SERIALIZER_CLASS_CONFIG, keySerializer);
        configProps.put(ProducerConfig.VALUE_SERIALIZER_CLASS_CONFIG, valueSerializer);

        Map<String, Object> sslConfigProps = getKafkaSSLConfigProps(tlsManager,keystoreLocation, truststoreLocation);
        if (sslConfigProps != null) {
            configProps.putAll(sslConfigProps);
        }
        return configProps;
    }

    private Map<String, Object> getKafkaConsumerConfigProps(TLSManager tlsManager,String groupId, String autoOffsetReset, String enableAutoCommit, String bootstrapServers, String keyDeserializer, String valueDeserializer, String keystoreLocation, String truststoreLocation) {
        Map<String, Object> configProps = new HashMap<>();
        configProps.put(ConsumerConfig.GROUP_ID_CONFIG, groupId);
        configProps.put(ConsumerConfig.AUTO_OFFSET_RESET_CONFIG, autoOffsetReset);
        configProps.put(ConsumerConfig.ENABLE_AUTO_COMMIT_CONFIG, enableAutoCommit);
        configProps.put(ConsumerConfig.BOOTSTRAP_SERVERS_CONFIG, bootstrapServers);
        configProps.put(ConsumerConfig.KEY_DESERIALIZER_CLASS_CONFIG, keyDeserializer);
        configProps.put(ConsumerConfig.VALUE_DESERIALIZER_CLASS_CONFIG, valueDeserializer);

        Map<String, Object> sslConfigProps = getKafkaSSLConfigProps(tlsManager,keystoreLocation, truststoreLocation);
        if (sslConfigProps != null) {
            configProps.putAll(sslConfigProps);
        }
        return configProps;
    }

    private Map<String, Object> getKafkaSSLConfigProps(TLSManager tlsManager,String keystoreLocation, String truststoreLocation) {
        try {
            return tlsManager.getKafkaSSLConfProps(keystoreLocation, truststoreLocation);
        } catch (Exception exc) {
            throw new RuntimeException("Could not configure Kafka with TLS configuration", exc);
        }
    }
}
