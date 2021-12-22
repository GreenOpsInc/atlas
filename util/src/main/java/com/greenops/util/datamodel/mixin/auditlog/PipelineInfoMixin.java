package com.greenops.util.datamodel.mixin.auditlog;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.List;

public class PipelineInfoMixin {
    @JsonProperty(value = "pipelineUniqueVersionNumber")
    private String pipelineUvn;

    @JsonProperty(value = "errors")
    private List<String> errors;

    @JsonProperty(value = "stepList")
    private List<String> stepList;

    @JsonCreator
    PipelineInfoMixin(@JsonProperty(value = "pipelineUniqueVersionNumber") String pipelineUvn,
                      @JsonProperty(value = "errors") List<String> errors,
                      @JsonProperty(value = "stepList") List<String> stepList) {
    }
}
