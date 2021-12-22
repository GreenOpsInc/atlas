package com.greenops.util.datamodel.auditlog;

import java.util.List;

public class PipelineInfo {

    private String pipelineUvn;
    private List<String> errors;
    private List<String> stepList;

    public PipelineInfo(String pipelineUvn, List<String> errors, List<String> stepList) {
        this.pipelineUvn = pipelineUvn;
        this.errors = errors;
        this.stepList = stepList;
    }

    public List<String> getErrors() {
        return errors;
    }

    public List<String> getStepList() {
        return stepList;
    }

    public void addError(String error) {
        this.errors.add(error);
    }

    public String getPipelineUvn() {
        return pipelineUvn;
    }
}
