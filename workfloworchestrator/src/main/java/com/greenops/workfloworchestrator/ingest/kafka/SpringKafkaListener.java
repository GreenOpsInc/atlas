package com.greenops.workfloworchestrator.ingest.kafka;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.util.datamodel.event.Event;
import com.greenops.util.error.AtlasNonRetryableError;
import com.greenops.workfloworchestrator.ingest.handling.EventHandler;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.kafka.support.Acknowledgment;
import org.springframework.stereotype.Component;

@Slf4j
@Component
public class SpringKafkaListener {

    @Autowired
    EventHandler eventHandler;

    @Autowired
    @Qualifier("eventAndRequestObjectMapper")
    ObjectMapper objectMapper;

    @KafkaListener(topics = "${spring.kafka.topic}", groupId = "${spring.kafka.consumer.group-id}")
    public void listen(String message, Acknowledgment ack) {
        try {
            var event = objectMapper.readValue(message, Event.class);
            eventHandler.handleEvent(event);
            ack.acknowledge();
        } catch (JsonProcessingException e) {
            log.error("ObjectMapper could not map message to Event", e);
            throw new AtlasNonRetryableError(e);
        }
    }

    //TODO: Add repartitioning handling
}
