package com.greenops.verificationtool.datamodel.pipelinedata;

import org.junit.jupiter.api.Test;

import java.util.List;

import static org.junit.jupiter.api.Assertions.assertEquals;

public class PipelineDataTest {

    @Test
    void assertThatPipelineDataSetsStepClusterNameCorrectly() {
        var pipelineClusterName = "cluster_us_east_2";
        var step2ClusterName = "cluster_us_west_2";
        var step1Name = "Step_1";
        var step2Name = "Step_2";
        var stepData1 = (StepData) new StepDataImpl(step1Name, null, null, null, null, null, 0, 0, null);
        var stepData2 = (StepData) new StepDataImpl(step2Name, null, null, null, step2ClusterName, null, 0, 0, null);
        var stepDataList = List.of(stepData1, stepData2);
        var pipelineData = new PipelineDataImpl("Pipeline_1", pipelineClusterName, false, stepDataList);
        assertEquals(pipelineData.getClusterName(), pipelineClusterName);
        assertEquals(pipelineData.getStep(step1Name).getClusterName(), pipelineClusterName);
        assertEquals(pipelineData.getStep(step2Name).getClusterName(), step2ClusterName);
    }
}
