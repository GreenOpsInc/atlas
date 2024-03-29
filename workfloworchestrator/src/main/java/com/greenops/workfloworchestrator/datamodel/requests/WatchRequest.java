package com.greenops.workfloworchestrator.datamodel.requests;

public class WatchRequest {

    private final String teamName;
    private final String pipelineName;
    private final String stepName;
    private final String type;
    private final String name;
    private final String namespace;
    private final int testNumber;

    public WatchRequest(String teamName, String pipelineName, String stepName, String type, String name, String namespace) {
        this(teamName, pipelineName, stepName, type, name, namespace, -1);
    }

    public WatchRequest(String teamName, String pipelineName, String stepName, String type, String name, String namespace, int testNumber) {
        this.teamName = teamName;
        this.pipelineName = pipelineName;
        this.stepName = stepName;
        this.type = type;
        this.name = name;
        this.namespace = namespace;
        this.testNumber = testNumber;
    }
}
