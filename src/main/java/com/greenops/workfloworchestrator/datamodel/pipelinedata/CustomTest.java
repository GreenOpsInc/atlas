package com.greenops.workfloworchestrator.datamodel.pipelinedata;

import java.util.Map;

public class CustomTest implements Test {

    private String path;
    private boolean executeInApplicationPod;
    private boolean executeBeforeDeployment;
    private Map<String, String> variables;

    CustomTest(String path, boolean executeInApplicationPod, boolean executeBeforeDeployment, Map<String, String> variables) {
        this.path = path;
        this.executeInApplicationPod = executeInApplicationPod;
        this.executeBeforeDeployment = executeBeforeDeployment;
        this.variables = variables;
    }
}
