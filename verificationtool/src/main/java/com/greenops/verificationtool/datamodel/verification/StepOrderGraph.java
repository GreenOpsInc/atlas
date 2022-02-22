package com.greenops.verificationtool.datamodel.verification;

import com.greenops.util.datamodel.pipelinedata.PipelineData;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;

public class StepOrderGraph {
    private final PipelineData pipelineData;
    private final HashMap<String, List<String>> graph;
    private final HashMap<String, Boolean> visited;

    public StepOrderGraph(PipelineData pipelineData) {
        this.pipelineData = pipelineData;
        this.graph = new HashMap<String, List<String>>();
        this.visited = new HashMap<String, Boolean>();
        this.createGraph();
    }

    private void createGraph() {
        var allStepNames = this.pipelineData.getAllSteps();
        for (String step : allStepNames) {
            List<String> childs = this.pipelineData.getChildrenSteps(step);
            for (String child : childs) {
                this.addEdge(step, child);
            }
        }
    }

    public HashMap<String, List<String>> getStepDependencyGraph() {
        return this.graph;
    }


    private void addEdge(String u, String v) {
        if (!this.graph.containsKey(u)) {
            this.graph.put(u, new ArrayList<String>());
        }
        this.graph.get(u).add(v);
    }
}