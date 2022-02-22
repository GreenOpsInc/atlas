package com.greenops.workfloworchestrator.ingest;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.dataformat.yaml.YAMLFactory;
import com.greenops.util.datamodel.mixin.pipelinedata.InjectScriptTestMixin;
import com.greenops.util.datamodel.mixin.pipelinedata.PipelineDataMixin;
import com.greenops.util.datamodel.mixin.pipelinedata.StepDataMixin;
import com.greenops.util.datamodel.pipelinedata.InjectScriptTest;
import com.greenops.util.datamodel.pipelinedata.PipelineData;
import com.greenops.util.datamodel.pipelinedata.PipelineDataImpl;
import com.greenops.util.datamodel.pipelinedata.StepDataImpl;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

public class DataSerializationTest {

    private ObjectMapper yamlObjectMapper;
    private ObjectMapper objectMapper;

    @BeforeEach
    void beforeEach() {
        yamlObjectMapper = new ObjectMapper(new YAMLFactory());
        objectMapper = new ObjectMapper()
                .addMixIn(PipelineDataImpl.class, PipelineDataMixin.class)
                .addMixIn(StepDataImpl.class, StepDataMixin.class)
                .addMixIn(InjectScriptTest.class, InjectScriptTestMixin.class);
    }

    //TODO: Obviously an incomplete test. This should be manually tested if the data model is updated.
    @Test
    void testPipelineSerialization() throws JsonProcessingException {
        var pipelineData = "\n" +
                "name: test_pipeline_schema\n" +
                "steps:\n" +
                "- name: deploy_to_dev\n" +
                "  argo_application: guestbook\n" +
                "  application_path: relative/to/root\n" +
                "- name: deploy_to_int\n" +
                "  argo_application: guestbook\n" +
                "  application_path: relative/to/root #not required; default is nil, will check ArgoCD for existing application\n" +
                "  additional_deployments: relative/to/root #not required; for pieces like Istio\n" +
                "  rollback: true #not required; default is false\n" +
                "  tests: #not required; executed sequentially\n" +
                "  - path: \"/relative/to/root\"\n" +
                "    type: inject\n" +
                "    in_application_pod: true\n" +
                "    before: true\n" +
                "    variables:\n" +
                "      <variable_name>: xyz\n" +
                "      <variable_name_2>: abc\n" +
                "  dependencies: #default is none, outlines steps that have to be completed before current step\n" +
                "  - deploy_to_dev";
        var pipelineObject = objectMapper.readValue(objectMapper.writeValueAsString(yamlObjectMapper.readValue(pipelineData, Object.class)), PipelineData.class);
    }
}
