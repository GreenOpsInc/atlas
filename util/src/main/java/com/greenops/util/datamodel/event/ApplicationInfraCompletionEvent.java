package com.greenops.util.datamodel.event;

import org.apache.logging.log4j.util.Strings;

import java.util.List;

public class ApplicationInfraCompletionEvent implements Event {

    private String orgName;
    private String teamName;
    private String pipelineName;
    private String pipelineUvn;
    private String stepName;
    private boolean success;
    private int deliveryAttempt;

    public ApplicationInfraCompletionEvent(String orgName, String teamName, String pipelineName, String pipelineUvn, String stepName, boolean success) {
        this.orgName = orgName;
        this.teamName = teamName;
        this.pipelineName = pipelineName;
        this.pipelineUvn = pipelineUvn;
        this.stepName = stepName;
        this.success = success;
    }

    @Override
    public String getOrgName() {
        return orgName;
    }

    @Override
    public String getTeamName() {
        return teamName;
    }

    @Override
    public String getPipelineName() {
        return pipelineName;
    }

    @Override
    public String getPipelineUvn() {
        return pipelineUvn;
    }

    @Override
    public String getStepName() {
        return stepName;
    }

    @Override
    public String getMQKey() {
        return Strings.join(List.of(pipelineUvn, stepName, ApplicationInfraCompletionEvent.class.getName(), getDeliveryAttempt()), '-');
    }

    @Override
    public void setDeliveryAttempt(int attempt) {
        this.deliveryAttempt = attempt;
    }

    @Override
    public int getDeliveryAttempt() {
        return deliveryAttempt;
    }

    public boolean isSuccess() {
        return success;
    }
}