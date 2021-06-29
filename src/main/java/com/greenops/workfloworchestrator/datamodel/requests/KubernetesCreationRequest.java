package com.greenops.workfloworchestrator.datamodel.requests;

import java.util.List;
import java.util.Map;

public class KubernetesCreationRequest {

    private final String kind;
    private final String objectName;
    private final String namespace;
    private final String imageName;
    private final List<String> command;
    private final List<String> args;
    private final Map<String, String> variables;

    public KubernetesCreationRequest(String kind, String objectName, String namespace, String imageName, List<String> command, List<String> args, Map<String, String> variables) {
        this.kind = kind;
        this.objectName = objectName;
        this.namespace = namespace;
        this.imageName = imageName;
        this.command = command;
        this.args = args;
        this.variables = variables;
    }
}
