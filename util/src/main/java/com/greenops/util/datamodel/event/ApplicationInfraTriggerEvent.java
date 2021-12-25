package com.greenops.util.datamodel.event;

import org.apache.logging.log4j.util.Strings;

import java.util.List;

public class ApplicationInfraTriggerEvent implements Event {

    private String orgName;
    private String teamName;
    private String pipelineName;
    private String pipelineUvn;
    private String stepName;
    private int deliveryAttempt;

    public ApplicationInfraTriggerEvent(String orgName, String teamName, String pipelineName, String pipelineUvn, String stepName) {
        this.orgName = orgName;
        this.teamName = teamName;
        this.pipelineName = pipelineName;
        this.pipelineUvn = pipelineUvn;
        this.stepName = stepName;
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
    public String getMQKey() {
        return Strings.join(List.of(pipelineUvn, stepName, ApplicationInfraTriggerEvent.class.getName(), getDeliveryAttempt()), '-');
    }

    @Override
    public void setDeliveryAttempt(int attempt) {
        this.deliveryAttempt = attempt;
    }

    @Override
    public int getDeliveryAttempt() {
        return deliveryAttempt;
    }

    @Override
    public String getStepName() {
        return stepName;
    }
}
