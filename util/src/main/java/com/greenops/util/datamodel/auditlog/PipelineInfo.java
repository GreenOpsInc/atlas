package com.greenops.util.datamodel.auditlog;

import java.util.List;

public class PipelineInfo {

    private String pipelineUvn;
    private List<String> errors;

    public PipelineInfo(String pipelineUvn, List<String> errors) {
        this.pipelineUvn = pipelineUvn;
        this.errors = errors;
    }

    public List<String> getErrors() {
        return errors;
    }

    public void addError(String error) {
        this.errors.add(error);
    }

    public String getPipelineUvn() {
        return pipelineUvn;
    }
}
