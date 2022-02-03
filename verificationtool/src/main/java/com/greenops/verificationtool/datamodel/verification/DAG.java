package com.greenops.verificationtool.datamodel.verification;

import com.greenops.util.datamodel.event.*;
import com.greenops.util.datamodel.pipelinedata.PipelineData;
import com.greenops.util.datamodel.pipelinedata.Test;


import java.util.*;

public class DAG {
    private final String ATLAS_ROOT_DATA = "ATLAS_ROOT_DATA";
    private final String PipelineTriggerEvent = "PipelineTriggerEvent";
    private final String PipelineCompletionEvent = "PipelineCompletionEvent";
    private final String TriggerStepEvent = "TriggerStepEvent";
    private final String ApplicationInfraCompletionEvent = "ApplicationInfraCompletionEvent";
    private final String ApplicationInfraTriggerEvent = "ApplicationInfraTriggerEvent";
    private final String ClientCompletionEvent = "ClientCompletionEvent";
    private final String TestCompletionEvent = "TestCompletionEvent";
    private final PipelineData pipelineData;
    private final String pipelineName;
    private final HashMap<Vertex, List<Vertex>> graph;
    private final HashMap<Vertex, List<Vertex>> reverseGraph;
    private Vertex root;
    private HashMap<Vertex, Boolean> visited;
    private List<Vertex> ancestorsList;

    public DAG(PipelineData pipelineData, String pipelineName) {
        this.pipelineData = pipelineData;
        this.pipelineName = pipelineName;
        this.graph = new HashMap<Vertex, List<Vertex>>();
        this.reverseGraph = new HashMap<Vertex, List<Vertex>>();
        this.visited = new HashMap<Vertex, Boolean>();

        this.createDAG();
    }

    private void createDAG() {
        Vertex pipelineTriggerEventVertex = new Vertex(this.PipelineTriggerEvent, this.pipelineName, this.ATLAS_ROOT_DATA);
        this.root = pipelineTriggerEventVertex;
        this.graph.put(pipelineTriggerEventVertex, new ArrayList<Vertex>());

        var allStepNames = this.pipelineData.getAllSteps();
        var rootSteps = this.getRootSteps(allStepNames);

        StepOrderGraph stepOrderGraph = new StepOrderGraph(this.pipelineData);
        var stepDependencyGraph = stepOrderGraph.getStepDependencyGraph();
        Vertex lastVertex = null;
        for (String rootStep : rootSteps) {
            Queue<Pair> queue = new LinkedList<Pair>();
            HashMap<String, Boolean> visited = new HashMap<String, Boolean>();

            visited.put(rootStep, true);
            queue.add(new Pair(rootStep, pipelineTriggerEventVertex));

            while (queue.size() != 0) {
                var currentNode = queue.poll();
                var nextVertex = this.addVerticesForStep(currentNode.getStep(), currentNode.getVertex());
                lastVertex = nextVertex;

                if (stepDependencyGraph.get(currentNode.getStep()) == null) {
                    continue;
                }

                for (String nextStep : stepDependencyGraph.get(currentNode.getStep())) {
                    if (visited.get(nextStep) == null) {
                        visited.put(nextStep, true);
                        queue.add(new Pair(nextStep, nextVertex));
                    }
                }
            }
        }
        Vertex pipelineCompletionVertex = new Vertex(this.PipelineCompletionEvent, this.pipelineName, this.ATLAS_ROOT_DATA);
        this.addEdge(lastVertex, pipelineCompletionVertex);
    }

    public HashMap<Vertex, List<Vertex>> getDAG() {
        return this.reverseGraph;
    }

    private List<String> getRootSteps(List<String> allStepNames) {
        var rootSteps = new ArrayList<String>();
        for (String step : allStepNames) {
            if (this.pipelineData.getParentSteps(step).get(0).equals(this.ATLAS_ROOT_DATA)) {
                rootSteps.add(step);
            }
        }
        return rootSteps;
    }

    private Vertex addVerticesForStep(String stepName, Vertex parent) {
        var stepData = this.pipelineData.getStep(stepName);
        var tests = stepData.getTests();

        Vertex triggerStepEventVertex = new Vertex(this.TriggerStepEvent, this.pipelineName, stepName);
        Vertex applicationInfraTriggerEventVertex = new Vertex(this.ApplicationInfraTriggerEvent, this.pipelineName, stepName);
        Vertex applicationInfraCompletionEventVertex = new Vertex(this.ApplicationInfraCompletionEvent, this.pipelineName, stepName);
        Vertex clientCompletionEventVertex = new Vertex(this.ClientCompletionEvent, this.pipelineName, stepName);

        this.addEdge(parent, triggerStepEventVertex);
        Vertex testCompletionEventVertex = null;
        Integer testNumber = 0;
        for (Test test : tests) {
            if (test.shouldExecuteBefore()) {
                if (testCompletionEventVertex == null) {
                    testCompletionEventVertex = new Vertex(this.TestCompletionEvent, this.pipelineName, stepName, testNumber);
                    this.addEdge(triggerStepEventVertex, testCompletionEventVertex);
                } else {
                    Vertex tempTestCompletionEventVertex = new Vertex(this.TestCompletionEvent, this.pipelineName, stepName, testNumber);
                    this.addEdge(testCompletionEventVertex, tempTestCompletionEventVertex);
                    testCompletionEventVertex = tempTestCompletionEventVertex;
                }
            }
            testNumber += 1;
        }
        if (testCompletionEventVertex == null) {
            this.addEdge(triggerStepEventVertex, applicationInfraTriggerEventVertex);
        } else {
            this.addEdge(testCompletionEventVertex, applicationInfraTriggerEventVertex);
        }
        testCompletionEventVertex = null;
        this.addEdge(applicationInfraTriggerEventVertex, applicationInfraCompletionEventVertex);
        this.addEdge(applicationInfraCompletionEventVertex, clientCompletionEventVertex);

        testNumber = 0;
        for (Test test : tests) {
            if (!test.shouldExecuteBefore()) {
                if (testCompletionEventVertex == null) {
                    testCompletionEventVertex = new Vertex(this.TestCompletionEvent, this.pipelineName, stepName, testNumber);
                    this.addEdge(clientCompletionEventVertex, testCompletionEventVertex);
                } else {
                    Vertex tempTestCompletionEventVertex = new Vertex(this.TestCompletionEvent, this.pipelineName, stepName, testNumber);
                    this.addEdge(testCompletionEventVertex, tempTestCompletionEventVertex);
                    testCompletionEventVertex = tempTestCompletionEventVertex;
                }
            }
            testNumber += 1;
        }

        if (testCompletionEventVertex == null) {
            return clientCompletionEventVertex;
        } else {
            return testCompletionEventVertex;
        }
    }

    public Boolean checkEventOrderInDag(Event event) {
        var vertex = createVertex(event);

        if (this.reverseGraph.get(vertex) == null) {
            if (vertex.getEventType().equals(this.PipelineTriggerEvent)) {
                this.visited.put(vertex, true);
                return true;
            } else {
                return false;
            }
        }
        var parents = this.reverseGraph.get(vertex);
        for (Vertex parent : parents) {
            if (!this.visited.containsKey(parent)) {
                return false;
            }
        }
        this.visited.put(vertex, true);
        return true;
    }

    public Boolean isLastVertexInDAG(Event event) {
        var vertex = this.createVertex(event);
        if(this.graph.get(vertex) != null){
            return this.graph.get(vertex).get(0).getEventType().equals(this.PipelineCompletionEvent);
        }
        return false;
    }

    public Boolean isTestCompletionBefore(Vertex vertex) {
        var tests = this.pipelineData.getStep(vertex.getStepName()).getTests();
        var testNumber = vertex.getTestNumber();
        return tests.get(testNumber).shouldExecuteBefore();
    }

    private Vertex createVertex(Event event) {
        var pipelineName = event.getPipelineName();
        var stepName = event.getStepName();
        String eventType = null;
        Vertex vertex = null;
        if (event instanceof PipelineTriggerEvent) {
            eventType = this.PipelineTriggerEvent;
        } else if (event instanceof ClientCompletionEvent) {
            eventType = this.ClientCompletionEvent;
        } else if (event instanceof TestCompletionEvent) {
            eventType = this.TestCompletionEvent;
            vertex = new Vertex(eventType, pipelineName, stepName, ((TestCompletionEvent) event).getTestNumber());
        } else if (event instanceof ApplicationInfraTriggerEvent) {
            eventType = this.ApplicationInfraTriggerEvent;
        } else if (event instanceof ApplicationInfraCompletionEvent) {
            eventType = this.ApplicationInfraCompletionEvent;
        } else if (event instanceof TriggerStepEvent) {
            eventType = this.TriggerStepEvent;
        } else if (event instanceof PipelineCompletionEvent) {
            eventType = this.PipelineCompletionEvent;
        }
        if (vertex == null) {
            vertex = new Vertex(eventType, pipelineName, stepName);
        }
        return vertex;
    }

    public List<Vertex> getPreviousVertices(Event event) {
        var vertex = this.createVertex(event);
        return this.reverseGraph.get(vertex);
    }

    private void addEdge(Vertex u, Vertex v) {
        if (!this.graph.containsKey(u)) {
            this.graph.put(u, new ArrayList<Vertex>());
        }
        if (!this.reverseGraph.containsKey(v)) {
            this.reverseGraph.put(v, new ArrayList<Vertex>());
        }
        this.reverseGraph.get(v).add(u);
        this.graph.get(u).add(v);
    }
}