package com.greenops.workflowtrigger.kafka;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Component;

@Component
public class KafkaClient {

    private final String topic;
    private final KafkaTemplate<String, String> kafkaTemplate;

    @Autowired
    public KafkaClient(@Value("${spring.kafka.topic}") String topic, KafkaTemplate<String, String> kafkaTemplate) {
        this.topic = topic;
        this.kafkaTemplate = kafkaTemplate;
    }

    public void sendMessage(String data) {
        kafkaTemplate.send(topic, data);
        kafkaTemplate.flush();
    }
}
