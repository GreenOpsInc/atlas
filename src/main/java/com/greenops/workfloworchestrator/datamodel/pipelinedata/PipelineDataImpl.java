package com.greenops.workfloworchestrator.datamodel.pipelinedata;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class PipelineDataImpl implements PipelineData {

    private String name;
    private List<StepData> steps;
    private Map<String, List<String>> stepParents;
    private Map<String, List<String>> stepChildren;

    public PipelineDataImpl(String name, List<StepData> stepDataList) {
        this.name = name;
        this.steps = stepDataList;
        this.stepParents = new HashMap<>();
        this.stepChildren = new HashMap<>();
        for (var step : stepDataList) {
            stepParents.put(step.getName(), step.getDependencies());
            for (var parentStep : step.getDependencies()) {
                if (stepChildren.containsKey(parentStep)) {
                    var list = stepChildren.get(parentStep);
                    list.add(step.getName());
                    stepChildren.put(parentStep, list);
                } else {
                    stepChildren.put(parentStep, List.of(step.getName()));
                }
            }
        }
    }

    private PipelineDataImpl() {}
}
