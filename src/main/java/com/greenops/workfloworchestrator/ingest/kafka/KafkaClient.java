package com.greenops.workfloworchestrator.ingest.kafka;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Component;

@Component
public class KafkaClient {

    private final String normalTopic;
    private final String dlqTopic;
    private final KafkaTemplate<String, String> kafkaTemplate;

    @Autowired
    public KafkaClient(@Value("${spring.kafka.topic}") String topic, @Value("${spring.kafka.dlqtopic}") String dlqTopic, KafkaTemplate<String, String> kafkaTemplate) {
        this.normalTopic = topic;
        this.dlqTopic = dlqTopic;
        this.kafkaTemplate = kafkaTemplate;
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
