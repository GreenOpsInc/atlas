package com.greenops.workfloworchestrator.ingest.kafka;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.util.datamodel.event.Event;
import com.greenops.util.error.AtlasNonRetryableError;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Component;

import java.util.List;

@Component
public class KafkaClient {

    private final ObjectMapper objectMapper;
    private final String normalTopic;
    private final String dlqTopic;
    private final KafkaTemplate<String, String> kafkaTemplate;

    @Autowired
    public KafkaClient(@Qualifier("eventAndRequestObjectMapper") ObjectMapper objectMapper, @Value("${application.kafka.topic}") String topic, @Value("${application.kafka.dlqtopic}") String dlqTopic, KafkaTemplate<String, String> kafkaTemplate) {
        this.objectMapper = objectMapper;
        this.normalTopic = topic;
        this.dlqTopic = dlqTopic;
        this.kafkaTemplate = kafkaTemplate;
    }

    public void sendMessage(Event event) {
        try {
            kafkaTemplate.send(normalTopic, objectMapper.writeValueAsString(event));
        } catch (JsonProcessingException e) {
            throw new AtlasNonRetryableError(e);
        }
        kafkaTemplate.flush();
    }

    public void sendMessage(List<Event> events) {
        try {
            for (var event : events) {
                kafkaTemplate.send(normalTopic, objectMapper.writeValueAsString(event));
            }
        } catch (JsonProcessingException e) {
            throw new AtlasNonRetryableError(e);
        }
        kafkaTemplate.flush();
    }

    public void sendMessage(String data) {
        kafkaTemplate.send(normalTopic, data);
        kafkaTemplate.flush();
    }

    public void sendMessageToDlq(String data) {
        kafkaTemplate.send(dlqTopic, data);
        kafkaTemplate.flush();
    }
}
