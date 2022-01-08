package com.greenops.verificationtool.datamodel.pipelinedata;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

import static com.greenops.verificationtool.datamodel.pipelinedata.StepData.ROOT_STEP_NAME;

public class PipelineDataImpl implements PipelineData {

    private String name;
    private List<StepData> steps;
    private final String clusterName;
    private final boolean argoVersionLock;
    private final Map<String, List<String>> stepParents;
    private final Map<String, List<String>> stepChildren;

    public PipelineDataImpl(String name, String clusterName, boolean argoVersionLock, List<StepData> stepDataList) {
        this.name = name;
        this.steps = stepDataList;
        this.clusterName = clusterName;
        this.argoVersionLock = argoVersionLock;
        this.stepParents = new HashMap<>();
        this.stepChildren = new HashMap<>();
        for (var step : stepDataList) {
            if (step.getClusterName() == null) {
                step.setClusterName(clusterName);
            }
            var stepDependencies = step.getDependencies();
            if (stepDependencies.isEmpty()) {
                var parents = new ArrayList<String>();
                parents.add(ROOT_STEP_NAME);
                stepParents.put(step.getName(), parents);
                stepDependencies.add(ROOT_STEP_NAME);
            } else {
                stepParents.put(step.getName(), stepDependencies);
            }
            for (var parentStep : stepDependencies) {
                if (stepChildren.containsKey(parentStep)) {
                    var list = stepChildren.get(parentStep);
                    list.add(step.getName());
                    stepChildren.put(parentStep, list);
                } else {
                    var children = new ArrayList<String>();
                    children.add(step.getName());
                    stepChildren.put(parentStep, children);
                }
            }
        }
    }

    @Override
    public String getName() {
        return name;
    }

    @Override
    public String getClusterName() {
        return clusterName;
    }

    @Override
    public StepData getStep(String stepName) {
        var stepMatches = steps.stream().filter(stepData -> stepData.getName().equals(stepName)).collect(Collectors.toList());
        return stepMatches.size() == 1 ? stepMatches.get(0) : null;
    }

    @Override
    public List<String> getChildrenSteps(String stepName) {
        return stepChildren.getOrDefault(stepName, List.of());
    }

    @Override
    public List<String> getParentSteps(String stepName) {
        return stepParents.getOrDefault(stepName, List.of());
    }

    @Override
    public List<String> getAllSteps() {
        return steps.stream().map(StepData::getName).collect(Collectors.toList());
    }

    @Override
    public List<String> getAllStepsOrdered() {
        var orderedSteps = new ArrayList<>(getChildrenSteps(ROOT_STEP_NAME));
        var stepsMap = getAllSteps().stream().collect(Collectors.toMap(step -> step, orderedSteps::contains));
        var idx = 0;
        while (idx < orderedSteps.size()) {
            var childrenList = getChildrenSteps(orderedSteps.get(idx));
            for (var childStep : childrenList) {
                var parentsSeen = true;
                for (var parentOfChildStep : getParentSteps(childStep)) {
                    if (!stepsMap.get(parentOfChildStep)) {
                        parentsSeen = false;
                        break;
                    }
                }
                if (parentsSeen && !stepsMap.get(childStep)) {
                    orderedSteps.add(childStep);
                    stepsMap.put(childStep, true);
                }
            }
            idx++;
        }
        return orderedSteps;
    }

    @Override
    public boolean isArgoVersionLock() {
        return argoVersionLock;
    }
}
