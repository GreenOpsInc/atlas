package com.greenops.workfloworchestrator.datamodel.pipelinedata;

import com.greenops.workfloworchestrator.datamodel.requests.KubernetesCreationRequest;

import java.util.Map;

import static com.greenops.workfloworchestrator.ingest.handling.ClientKey.makeTestKey;
import static com.greenops.workfloworchestrator.ingest.handling.util.deployment.SchemaHandlingUtil.getFileNameWithoutExtension;

public class CustomJobTest implements Test {

    private String path;
    private boolean executeBeforeDeployment;
    private Map<String, String> variables;

    CustomJobTest(String path, boolean executeBeforeDeployment, Map<String, String> variables) {
        this.path = path;
        this.executeBeforeDeployment = executeBeforeDeployment;
        this.variables = variables;
    }

    @Override
    public String getPath() {
        return path;
    }

    @Override
    public boolean shouldExecuteBefore() {
        return executeBeforeDeployment;
    }

    @Override
    public Map<String, String> getVariables() {
        return variables;
    }

    @Override
    public Object getPayload(int testNumber, String testConfig, String teamName, String pipelineName, String stepName) {
        return new KubernetesCreationRequest(makeTestKey(teamName, pipelineName, stepName, getFileNameWithoutExtension(getPath())), testConfig, getVariables());
    }
}
