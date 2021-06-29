package com.greenops.workfloworchestrator.datamodel.requests;

import java.util.List;
import java.util.Map;

public class WatchRequest {

    private final String teamName;
    private final String pipelineName;
    private final String stepName;
    private final String type;
    private final String name;
    private final String namespace;

    public WatchRequest(String teamName, String pipelineName, String stepName, String type, String name, String namespace) {
        this.teamName = teamName;
        this.pipelineName = pipelineName;
        this.stepName = stepName;
        this.type = type;
        this.name = name;
        this.namespace = namespace;
    }
}
