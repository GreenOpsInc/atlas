package com.greenops.verificationtool.datamodel.verification;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.util.datamodel.pipelinedata.PipelineData;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;

import static org.junit.jupiter.api.Assertions.assertEquals;

@SpringBootTest
public class DagTest {
    private final String ATLAS_ROOT_DATA =  "ATLAS_ROOT_DATA";
    private final String PipelineTriggerEvent = "PipelineTriggerEvent";
    private final String PipelineCompletionEvent = "PipelineCompletionEvent";
    private final String TriggerStepEvent = "TriggerStepEvent";
    private final String ApplicationInfraCompletionEvent = "ApplicationInfraCompletionEvent";
    private final String ApplicationInfraTriggerEvent = "ApplicationInfraTriggerEvent";
    private final String ClientCompletionEvent = "ClientCompletionEvent";
    private final String TestCompletionEvent = "TestCompletionEvent";
    private final String pipelineName = "examplePipeline";
    @Autowired
    private ObjectMapper yamlObjectMapper;
    @Autowired
    private ObjectMapper objectMapper;

    private HashMap<Vertex, List<Vertex>> getDAG(String yamlPipeline) {
        PipelineData pipelineObj = null;
        try {
            pipelineObj = objectMapper.readValue(
                    objectMapper.writeValueAsString(
                            yamlObjectMapper.readValue(yamlPipeline, Object.class)
                    ),
                    PipelineData.class);
        } catch (JsonProcessingException e) {
            throw new RuntimeException(e.getMessage());
        }
        DAG dag = new DAG(pipelineObj, this.pipelineName);
        return dag.getDAG();
    }

    @Test
    void assertThatDagConstructedCorrectlyUsingPipelineOne() {
        String pipelineYaml = "name: examplePipeline\n" +
                "argo_version_lock: true\n" +
                "cluster_name: kubernetes_local\n" +
                "steps:\n" +
                "- name: deploy_to_dev\n" +
                "  application_path: testapp_dev.yml\n" +
                "  tests:\n" +
                "  - path: verifyendpoints.sh\n" +
                "    type: inject\n" +
                "    image: alpine:latest\n" +
                "    commands: [sh, -c, ./verifyendpoints.sh]\n" +
                "    before: false\n" +
                "    variables:\n" +
                "      SERVICE_INTERNAL_URL: testapp.dev.svc.cluster.local";

        HashMap<Vertex, List<Vertex>> graph = getDAG(pipelineYaml);

        String pipelineName = "examplePipeline";
        String stepName = "deploy_to_dev";
        Vertex pipelineTriggerEvent = new Vertex(this.PipelineTriggerEvent, pipelineName, ATLAS_ROOT_DATA);
        Vertex triggerStepEvent = new Vertex(this.TriggerStepEvent, pipelineName, stepName);
        Vertex applicationInfraCompletionEvent = new Vertex(this.ApplicationInfraCompletionEvent, pipelineName, stepName);
        Vertex applicationInfraTriggerEvent = new Vertex(this.ApplicationInfraTriggerEvent, pipelineName, stepName);
        Vertex clientCompletionEvent = new Vertex(this.ClientCompletionEvent, pipelineName, stepName);
        Vertex testCompletionEvent = new Vertex(this.TestCompletionEvent, pipelineName, stepName, 0);

        assertEquals(graph.containsKey(pipelineTriggerEvent), false);
        assertEquals(graph.get(triggerStepEvent), List.of(pipelineTriggerEvent));
        assertEquals(graph.get(applicationInfraTriggerEvent), List.of(triggerStepEvent));
        assertEquals(graph.get(applicationInfraCompletionEvent), List.of(applicationInfraTriggerEvent));
        assertEquals(graph.get(clientCompletionEvent), List.of(applicationInfraCompletionEvent));
        assertEquals(graph.get(testCompletionEvent), List.of(clientCompletionEvent));
    }

    @Test
    void assertThatDagConstructedCorrectlyUsingPipelineTwo() {
        String pipelineYaml = "name: examplePipeline\n" +
                "argo_version_lock: true\n" +
                "cluster_name: kubernetes_local\n" +
                "steps:\n" +
                "  - name: deploy_to_dev\n" +
                "    application_path: testapp_dev.yml\n" +
                "    tests:\n" +
                "      - path: verifyendpoints.sh\n" +
                "        type: inject\n" +
                "        image: alpine:latest\n" +
                "        commands: [sh, -c, ./verifyendpoints.sh]\n" +
                "        before: true\n" +
                "        variables:\n" +
                "          SERVICE_INTERNAL_URL: testapp.dev.svc.cluster.local\n" +
                "      - path: verifyendpoints2.sh\n" +
                "        type: inject\n" +
                "        image: alpine:latest\n" +
                "        commands: [sh, -c, ./verifyendpoints2.sh]\n" +
                "        before: false\n" +
                "        variables:\n" +
                "          SERVICE_INTERNAL_URL: testapp.dev.svc.cluster.local\n" +
                "  - name: deploy_to_int\n" +
                "    application_path: testapp_dev.yml\n" +
                "    tests:\n" +
                "      - path: verifyendpoints.sh\n" +
                "        type: inject\n" +
                "        image: alpine:latest\n" +
                "        commands: [ sh, -c, ./verifyendpoints.sh ]\n" +
                "        before: false\n" +
                "        variables:\n" +
                "          SERVICE_INTERNAL_URL: testapp.dev.svc.cluster.local";

        HashMap<Vertex, List<Vertex>> graph = getDAG(pipelineYaml);

        String pipelineName = "examplePipeline";
        String firstStepName = "deploy_to_dev";
        String secondStepName = "deploy_to_int";
        Vertex pipelineTriggerEvent = new Vertex(this.PipelineTriggerEvent, pipelineName, ATLAS_ROOT_DATA);
        Vertex firstTriggerStepEvent = new Vertex(this.TriggerStepEvent, pipelineName, firstStepName);
        Vertex firstApplicationInfraCompletionEvent = new Vertex(this.ApplicationInfraCompletionEvent, pipelineName, firstStepName);
        Vertex firstApplicationInfraTriggerEvent = new Vertex(this.ApplicationInfraTriggerEvent, pipelineName, firstStepName);
        Vertex firstClientCompletionEvent = new Vertex(this.ClientCompletionEvent, pipelineName, firstStepName);
        Vertex firstTestCompletionEvent0 = new Vertex(this.TestCompletionEvent, pipelineName, firstStepName, 0);
        Vertex firstTestCompletionEvent1 = new Vertex(this.TestCompletionEvent, pipelineName, firstStepName, 1);

        Vertex secondTriggerStepEvent = new Vertex(this.TriggerStepEvent, pipelineName, secondStepName);
        Vertex secondApplicationInfraCompletionEvent = new Vertex(this.ApplicationInfraCompletionEvent, pipelineName, secondStepName);
        Vertex secondApplicationInfraTriggerEvent = new Vertex(this.ApplicationInfraTriggerEvent, pipelineName, secondStepName);
        Vertex secondClientCompletionEvent = new Vertex(this.ClientCompletionEvent, pipelineName, secondStepName);
        Vertex secondTestCompletionEvent = new Vertex(this.TestCompletionEvent, pipelineName, secondStepName, 0);

        assertEquals(graph.containsKey(pipelineTriggerEvent), false);
        assertEquals(graph.get(firstTriggerStepEvent), List.of(pipelineTriggerEvent));
        assertEquals(graph.get(firstTestCompletionEvent0), List.of(firstTriggerStepEvent));
        assertEquals(graph.get(firstApplicationInfraTriggerEvent), List.of(firstTestCompletionEvent0));
        assertEquals(graph.get(firstApplicationInfraCompletionEvent), List.of(firstApplicationInfraTriggerEvent));
        assertEquals(graph.get(firstClientCompletionEvent), List.of(firstApplicationInfraCompletionEvent));
        assertEquals(graph.get(firstTestCompletionEvent1), List.of(firstClientCompletionEvent));

        assertEquals(graph.get(secondTriggerStepEvent), List.of(pipelineTriggerEvent));
        assertEquals(graph.get(secondApplicationInfraTriggerEvent), List.of(secondTriggerStepEvent));
        assertEquals(graph.get(secondApplicationInfraCompletionEvent), List.of(secondApplicationInfraTriggerEvent));
        assertEquals(graph.get(secondClientCompletionEvent), List.of(secondApplicationInfraCompletionEvent));
        assertEquals(graph.get(secondTestCompletionEvent), List.of(secondClientCompletionEvent));
    }

    @Test
    void assertThatDagConstructedCorrectlyUsingPipelineThree() {
        String pipelineYaml = "name: examplePipeline\n" +
                "argo_version_lock: true\n" +
                "cluster_name: kubernetes_local\n" +
                "steps:\n" +
                "  - name: deploy_to_dev\n" +
                "    application_path: testapp_dev.yml\n" +
                "    tests:\n" +
                "      - path: verifyendpoints.sh\n" +
                "        type: inject\n" +
                "        image: alpine:latest\n" +
                "        commands: [sh, -c, ./verifyendpoints.sh]\n" +
                "        before: true\n" +
                "        variables:\n" +
                "          SERVICE_INTERNAL_URL: testapp.dev.svc.cluster.local\n" +
                "  - name: deploy_to_int\n" +
                "    application_path: testapp_dev.yml\n" +
                "    tests:\n" +
                "      - path: verifyendpoints.sh\n" +
                "        type: inject\n" +
                "        image: alpine:latest\n" +
                "        commands: [ sh, -c, ./verifyendpoints.sh ]\n" +
                "        before: false\n" +
                "        variables:\n" +
                "          SERVICE_INTERNAL_URL: testapp.dev.svc.cluster.local\n" +
                "  - name: deploy_to_stage\n" +
                "    application_path: testapp_dev.yml\n" +
                "    tests:\n" +
                "      - path: verifyendpoints.sh\n" +
                "        type: inject\n" +
                "        image: alpine:latest\n" +
                "        commands: [ sh, -c, ./verifyendpoints.sh ]\n" +
                "        before: false\n" +
                "        variables:\n" +
                "          SERVICE_INTERNAL_URL: testapp.dev.svc.cluster.local\n" +
                "    dependencies:\n" +
                "      - deploy_to_dev\n" +
                "      - deploy_to_int\n";

        HashMap<Vertex, List<Vertex>> graph = getDAG(pipelineYaml);

        String pipelineName = "examplePipeline";
        String firstStepName = "deploy_to_dev";
        String secondStepName = "deploy_to_int";
        String thirdStepName = "deploy_to_stage";
        Vertex pipelineTriggerEvent = new Vertex(this.PipelineTriggerEvent, pipelineName, "ATLAS_ROOT_DATA");
        Vertex firstTriggerStepEvent = new Vertex(this.TriggerStepEvent, pipelineName, firstStepName);
        Vertex firstApplicationInfraCompletionEvent = new Vertex(this.ApplicationInfraCompletionEvent, pipelineName, firstStepName);
        Vertex firstApplicationInfraTriggerEvent = new Vertex(this.ApplicationInfraTriggerEvent, pipelineName, firstStepName);
        Vertex firstClientCompletionEvent = new Vertex(this.ClientCompletionEvent, pipelineName, firstStepName);
        Vertex firstTestCompletionEvent = new Vertex(this.TestCompletionEvent, pipelineName, firstStepName, 0);

        Vertex secondTriggerStepEvent = new Vertex(this.TriggerStepEvent, pipelineName, secondStepName);
        Vertex secondApplicationInfraCompletionEvent = new Vertex(this.ApplicationInfraCompletionEvent, pipelineName, secondStepName);
        Vertex secondApplicationInfraTriggerEvent = new Vertex(this.ApplicationInfraTriggerEvent, pipelineName, secondStepName);
        Vertex secondClientCompletionEvent = new Vertex(this.ClientCompletionEvent, pipelineName, secondStepName);
        Vertex secondTestCompletionEvent = new Vertex(this.TestCompletionEvent, pipelineName, secondStepName, 0);

        Vertex thirdTriggerStepEvent = new Vertex(this.TriggerStepEvent, pipelineName, thirdStepName);
        Vertex thirdApplicationInfraCompletionEvent = new Vertex(this.ApplicationInfraCompletionEvent, pipelineName, thirdStepName);
        Vertex thirdApplicationInfraTriggerEvent = new Vertex(this.ApplicationInfraTriggerEvent, pipelineName, thirdStepName);
        Vertex thirdClientCompletionEvent = new Vertex(this.ClientCompletionEvent, pipelineName, thirdStepName);
        Vertex thirdTestCompletionEvent = new Vertex(this.TestCompletionEvent, pipelineName, thirdStepName, 0);

        assertEquals(graph.containsKey(pipelineTriggerEvent), false);
        assertEquals(graph.get(firstTriggerStepEvent), List.of(pipelineTriggerEvent));
        assertEquals(graph.get(firstTestCompletionEvent), List.of(firstTriggerStepEvent));
        assertEquals(graph.get(firstApplicationInfraTriggerEvent), List.of(firstTestCompletionEvent));
        assertEquals(graph.get(firstApplicationInfraCompletionEvent), List.of(firstApplicationInfraTriggerEvent));
        assertEquals(graph.get(firstClientCompletionEvent), List.of(firstApplicationInfraCompletionEvent));

        assertEquals(graph.get(secondTriggerStepEvent), List.of(pipelineTriggerEvent));
        assertEquals(graph.get(secondApplicationInfraTriggerEvent), List.of(secondTriggerStepEvent));
        assertEquals(graph.get(secondApplicationInfraCompletionEvent), List.of(secondApplicationInfraTriggerEvent));
        assertEquals(graph.get(secondClientCompletionEvent), List.of(secondApplicationInfraCompletionEvent));
        assertEquals(graph.get(secondTestCompletionEvent), List.of(secondClientCompletionEvent));

        List<Vertex> thirdTriggerStepEventParents = new ArrayList<Vertex>();
        thirdTriggerStepEventParents.add(firstClientCompletionEvent);
        thirdTriggerStepEventParents.add(secondTestCompletionEvent);
        assertEquals(graph.get(thirdTriggerStepEvent), thirdTriggerStepEventParents);
        assertEquals(graph.get(thirdApplicationInfraTriggerEvent), List.of(thirdTriggerStepEvent, thirdTriggerStepEvent));
        assertEquals(graph.get(thirdApplicationInfraCompletionEvent), List.of(thirdApplicationInfraTriggerEvent, thirdApplicationInfraTriggerEvent));
        assertEquals(graph.get(thirdClientCompletionEvent), List.of(thirdApplicationInfraCompletionEvent, thirdApplicationInfraCompletionEvent));
        assertEquals(graph.get(thirdTestCompletionEvent), List.of(thirdClientCompletionEvent, thirdClientCompletionEvent));
    }

    @Test
    void assertThatDagConstructedCorrectlyUsingPipelineFour() {
        String pipelineYaml = "name: examplePipeline\n" +
                "argo_version_lock: true\n" +
                "cluster_name: in-cluster\n" +
                "steps:\n" +
                "- name: deploy_to_dev\n" +
                "  application_path: testapp_dev.yml\n" +
                "  tests:\n" +
                "  - path: verifyendpoints.sh\n" +
                "    type: inject\n" +
                "    image: curlimages/curl:latest\n" +
                "    commands: [sh, -c, ./verifyendpoints.sh]\n" +
                "    before: true\n" +
                "    variables:\n" +
                "      SERVICE_INTERNAL_URL: testapp.dev.svc.cluster.local\n" +
                "- name: deploy_to_int\n" +
                "  application_path: testapp_int.yml\n" +
                "  tests:\n" +
                "  - path: verifyendpoints.sh\n" +
                "    type: inject\n" +
                "    image: curlimages/curl:latest\n" +
                "    commands: [sh, -c, ./verifyendpoints.sh]\n" +
                "    before: false\n" +
                "    variables:\n" +
                "      SERVICE_INTERNAL_URL: testapp.int.svc.cluster.local\n" +
                "  dependencies:\n" +
                "  - deploy_to_dev\n" +
                "- name: deploy_to_saif\n" +
                "  application_path: testapp_dev.yml\n" +
                "  tests:\n" +
                "  - path: verifyendpoints.sh\n" +
                "    type: inject\n" +
                "    image: curlimages/curl:latest\n" +
                "    commands: [sh, -c, ./verifyendpoints.sh]\n" +
                "    before: false\n" +
                "    variables:\n" +
                "      SERVICE_INTERNAL_URL: testapp.dev.svc.cluster.local\n" +
                "  dependencies:\n" +
                "  - deploy_to_dev";

        HashMap<Vertex, List<Vertex>> graph = getDAG(pipelineYaml);

        String pipelineName = "examplePipeline";
        String firstStepName = "deploy_to_dev";
        String secondStepName = "deploy_to_int";
        String thirdStepName = "deploy_to_saif";
        Vertex pipelineTriggerEvent = new Vertex(this.PipelineTriggerEvent, pipelineName, ATLAS_ROOT_DATA);
        Vertex firstTriggerStepEvent = new Vertex(this.TriggerStepEvent, pipelineName, firstStepName);
        Vertex firstApplicationInfraCompletionEvent = new Vertex(this.ApplicationInfraCompletionEvent, pipelineName, firstStepName);
        Vertex firstApplicationInfraTriggerEvent = new Vertex(this.ApplicationInfraTriggerEvent, pipelineName, firstStepName);
        Vertex firstClientCompletionEvent = new Vertex(this.ClientCompletionEvent, pipelineName, firstStepName);
        Vertex firstTestCompletionEvent = new Vertex(this.TestCompletionEvent, pipelineName, firstStepName, 0);

        Vertex secondTriggerStepEvent = new Vertex(this.TriggerStepEvent, pipelineName, secondStepName);
        Vertex secondApplicationInfraCompletionEvent = new Vertex(this.ApplicationInfraCompletionEvent, pipelineName, secondStepName);
        Vertex secondApplicationInfraTriggerEvent = new Vertex(this.ApplicationInfraTriggerEvent, pipelineName, secondStepName);
        Vertex secondClientCompletionEvent = new Vertex(this.ClientCompletionEvent, pipelineName, secondStepName);
        Vertex secondTestCompletionEvent = new Vertex(this.TestCompletionEvent, pipelineName, secondStepName, 0);

        Vertex thirdTriggerStepEvent = new Vertex(this.TriggerStepEvent, pipelineName, thirdStepName);
        Vertex thirdApplicationInfraCompletionEvent = new Vertex(this.ApplicationInfraCompletionEvent, pipelineName, thirdStepName);
        Vertex thirdApplicationInfraTriggerEvent = new Vertex(this.ApplicationInfraTriggerEvent, pipelineName, thirdStepName);
        Vertex thirdClientCompletionEvent = new Vertex(this.ClientCompletionEvent, pipelineName, thirdStepName);
        Vertex thirdTestCompletionEvent = new Vertex(this.TestCompletionEvent, pipelineName, thirdStepName, 0);

        Vertex pipelineCompletionEvent = new Vertex(this.PipelineCompletionEvent, pipelineName, ATLAS_ROOT_DATA);

        assertEquals(graph.containsKey(pipelineTriggerEvent), false);
        assertEquals(graph.get(firstTriggerStepEvent), List.of(pipelineTriggerEvent));
        assertEquals(graph.get(firstTestCompletionEvent), List.of(firstTriggerStepEvent));
        assertEquals(graph.get(firstApplicationInfraTriggerEvent), List.of(firstTestCompletionEvent));
        assertEquals(graph.get(firstApplicationInfraCompletionEvent), List.of(firstApplicationInfraTriggerEvent));
        assertEquals(graph.get(firstClientCompletionEvent), List.of(firstApplicationInfraCompletionEvent));

        assertEquals(graph.get(secondTriggerStepEvent), List.of(firstClientCompletionEvent));
        assertEquals(graph.get(secondApplicationInfraTriggerEvent), List.of(secondTriggerStepEvent));
        assertEquals(graph.get(secondApplicationInfraCompletionEvent), List.of(secondApplicationInfraTriggerEvent));
        assertEquals(graph.get(secondClientCompletionEvent), List.of(secondApplicationInfraCompletionEvent));
        assertEquals(graph.get(secondTestCompletionEvent), List.of(secondClientCompletionEvent));

        assertEquals(graph.get(thirdTriggerStepEvent), List.of(firstClientCompletionEvent));
        assertEquals(graph.get(thirdApplicationInfraTriggerEvent), List.of(thirdTriggerStepEvent));
        assertEquals(graph.get(thirdApplicationInfraCompletionEvent), List.of(thirdApplicationInfraTriggerEvent));
        assertEquals(graph.get(thirdClientCompletionEvent), List.of(thirdApplicationInfraCompletionEvent));
        assertEquals(graph.get(thirdTestCompletionEvent), List.of(thirdClientCompletionEvent));

        List<Vertex> pipelineCompletionEventParents = new ArrayList<Vertex>();
        pipelineCompletionEventParents.add(firstClientCompletionEvent);
        pipelineCompletionEventParents.add(secondTestCompletionEvent);
        pipelineCompletionEventParents.add(thirdTestCompletionEvent);
        assertEquals(graph.get(pipelineCompletionEvent), pipelineCompletionEventParents);
    }
}