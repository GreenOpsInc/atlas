package com.greenops.verificationtool.api;

import java.io.StringReader;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import com.greenops.verificationtool.datamodel.pipelinedata.PipelineData;
import org.springframework.beans.factory.annotation.Qualifier;

@RestController
public class VerificationApi {

    private final ObjectMapper yamlObjectMapper;
    private final ObjectMapper objectMapper;

    @Autowired
    public VerificationApi(@Qualifier("objectMapper") ObjectMapper objectMapper, @Qualifier("yamlObjectMapper")  ObjectMapper yamlObjectMapper){
        this.objectMapper = objectMapper;
        this.yamlObjectMapper = yamlObjectMapper;
    }

    @PostMapping(value = "/verify/{pipelineName}")
    public ResponseEntity<String> verifyPipeline(@PathVariable("pipelineName") String pipelineName,
                                               @RequestBody String yamlPipeline){
        if(pipelineName == null){
            return ResponseEntity.badRequest().build();
        }

        try {
            System.out.println("YAML OBJECT" + objectMapper.writeValueAsString(
                    yamlObjectMapper.readValue(yamlPipeline, Object.class)
            ));
            PipelineData pipelineObj = objectMapper.readValue(
                    objectMapper.writeValueAsString(
                            yamlObjectMapper.readValue(yamlPipeline, Object.class)
                    ),
                    PipelineData.class);
            System.out.println("NAME" + pipelineObj.getName());
            return ResponseEntity.ok().body(schemaToResponsePayload(pipelineObj));
        } catch(JsonProcessingException e) {
            System.out.println("ERROR VERIFY" + e);
            return ResponseEntity.badRequest().build();
        }
    }

    private String schemaToResponsePayload(Object schema) {
        try {
            return objectMapper.writeValueAsString(schema);
        } catch (JsonProcessingException e) {
            throw new RuntimeException("Could not convert schema into response payload.", e);
        }
    }
}