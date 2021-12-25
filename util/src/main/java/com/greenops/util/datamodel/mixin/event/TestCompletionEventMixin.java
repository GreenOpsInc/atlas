package com.greenops.util.datamodel.mixin.event;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class TestCompletionEventMixin extends EventMixin {

    @JsonProperty(value = "successful")
    boolean successful;
    @JsonProperty(value = "log")
    String log;
    @JsonProperty(value = "testName")
    String testName;
    @JsonProperty(value = "testNumber")
    int testNumber;

    @JsonCreator
    public TestCompletionEventMixin(@JsonProperty(value = "successful") boolean successful,
                                    @JsonProperty(value = "orgName") String orgName,
                                    @JsonProperty(value = "teamName") String teamName,
                                    @JsonProperty(value = "pipelineName") String pipelineName,
                                    @JsonProperty(value = "pipelineUvn") String pipelineUvn,
                                    @JsonProperty(value = "stepName") String stepName,
                                    @JsonProperty(value = "log") String log,
                                    @JsonProperty(value = "testName") String testName,
                                    @JsonProperty(value = "testNumber") int testNumber) {
    }
}
