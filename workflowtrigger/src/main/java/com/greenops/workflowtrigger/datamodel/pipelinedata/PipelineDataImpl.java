package com.greenops.workflowtrigger.datamodel.pipelinedata;

import java.util.List;
import java.util.stream.Collectors;

public class PipelineDataImpl implements PipelineData {

    private List<StepData> steps;
    private final String clusterName;

    public PipelineDataImpl(String clusterName, List<StepData> stepDataList) {
        this.steps = stepDataList.stream().map(stepData -> {
            if (stepData.getClusterName() == null) {
                stepData.setClusterName(clusterName);
            }
            return stepData;
        }).collect(Collectors.toList());
        this.clusterName = clusterName;
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
    public List<String> getAllSteps() {
        return steps.stream().map(StepData::getName).collect(Collectors.toList());
    }
}
