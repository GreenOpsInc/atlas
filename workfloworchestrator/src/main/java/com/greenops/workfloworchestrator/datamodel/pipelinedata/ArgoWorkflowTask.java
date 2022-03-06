package com.greenops.workfloworchestrator.datamodel.pipelinedata;

import com.greenops.workfloworchestrator.datamodel.requests.KubernetesCreationRequest;

import java.util.List;

public class ArgoWorkflowTask implements Test {
    private String path;
    private boolean executeBeforeDeployment;
    private List<Object> variables;

    ArgoWorkflowTask(String path, boolean executeBeforeDeployment, List<Object> variables) {
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
        return new KubernetesCreationRequest(ARGO_WORKFLOW_TASK, testConfig, getVariables());
    }

    @Override
    public String getWatchKey() {
        return ARGO_WORKFLOW_TASK;
    }
}
