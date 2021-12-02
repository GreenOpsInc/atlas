package com.greenops.workfloworchestrator.datamodel.requests;

import java.util.List;
import java.util.Map;

public class KubernetesCreationRequest {

    private final String type;
    private final String kind;
    private final String objectName;
    private final String namespace;
    private final String imageName;
    private final List<String> command;
    private final List<String> args;
    private final String configPayload;
    private final String volumeFilename;
    private final String volumePayload;
    private final Map<String, String> variables;

    public KubernetesCreationRequest(String type, String configPayload, Map<String, String> variables) {
        this(type, null, null, null, null, null, null, configPayload, null, null, variables);
    }

    public KubernetesCreationRequest(String type, String kind, String objectName, String namespace, String imageName, List<String> command, List<String> args, String volumeFilename, String volumePayload, Map<String, String> variables) {
        this(type, kind, objectName, namespace, imageName, command, args, null, volumeFilename, volumePayload, variables);
    }

    //Mainly for the Mixin
    public KubernetesCreationRequest(String type, String kind, String objectName, String namespace, String imageName, List<String> command, List<String> args, String configPayload, String volumeFilename, String volumePayload, Map<String, String> variables) {
        this.type = type;
        this.kind = kind;
        this.objectName = objectName;
        this.namespace = namespace;
        this.imageName = imageName;
        this.command = command;
        this.args = args;
        this.configPayload = configPayload;
        this.volumeFilename = volumeFilename;
        this.volumePayload = volumePayload;
        this.variables = variables;
    }
}
