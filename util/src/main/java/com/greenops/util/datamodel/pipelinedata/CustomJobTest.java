package com.greenops.util.datamodel.pipelinedata;

import com.greenops.util.datamodel.request.KubernetesCreationRequest;

import java.util.Map;

public class CustomJobTest implements Test {
    public static final String WATCH_TEST_KEY = "WatchTestKey";
    private final String path;
    private final boolean executeBeforeDeployment;
    private final Map<String, String> variables;

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
    public Object getPayload(int testNumber, String testConfig) {
        return new KubernetesCreationRequest(CUSTOM_TASK, testConfig, getVariables());
    }

    @Override
    public String getWatchKey() {
        return WATCH_TEST_KEY;
    }
}
