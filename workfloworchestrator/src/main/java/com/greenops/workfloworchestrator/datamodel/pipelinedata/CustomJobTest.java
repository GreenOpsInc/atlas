package com.greenops.workfloworchestrator.datamodel.pipelinedata;

import com.greenops.workfloworchestrator.datamodel.requests.KubernetesCreationRequest;

import java.util.List;
import java.util.Map;

import static com.greenops.workfloworchestrator.ingest.handling.EventHandlerImpl.WATCH_TEST_KEY;

public class CustomJobTest implements Test {

    private String path;
    private boolean executeBeforeDeployment;
    private List<Object> variables;

    CustomJobTest(String path, boolean executeBeforeDeployment, List<Object> variables) {
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
    public List<Object> getVariables() {
        return variables;
    }

    @Override
    public Object getPayload(int testNumber, String testConfig) {
        return new KubernetesCreationRequest(CUSTOM_TASK, testConfig, getVariables());
    }

    @Override
    public String getWatchKey() {
        return WATCH_TEST_KEY;
    }
}
