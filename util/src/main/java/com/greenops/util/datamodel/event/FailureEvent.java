package com.greenops.util.datamodel.event;

import com.greenops.util.datamodel.request.DeployResponse;
import org.apache.logging.log4j.util.Strings;

import java.util.List;

public class FailureEvent implements Event {

    private String orgName;
    private String teamName;
    private String pipelineName;
    private String pipelineUvn;
    private String stepName;
    private DeployResponse deployResponse;
    private String statusCode;
    private String error;
    private int deliveryAttempt;

    public FailureEvent(String orgName, String teamName, String pipelineName, String pipelineUvn, String stepName, DeployResponse deployResponse, String statusCode, String error) {
        this.orgName = orgName;
        this.teamName = teamName;
        this.pipelineName = pipelineName;
        this.pipelineUvn = pipelineUvn;
        this.stepName = stepName;
        this.deployResponse = deployResponse;
        this.statusCode = statusCode;
        this.error = error;
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
        return Strings.join(List.of(pipelineUvn, stepName, FailureEvent.class.getName(), getDeliveryAttempt()), '-');
    }

    @Override
    public void setDeliveryAttempt(int attempt) {
        this.deliveryAttempt = attempt;
    }

    @Override
    public int getDeliveryAttempt() {
        return deliveryAttempt;
    }

    public DeployResponse getDeployResponse() {
        return deployResponse;
    }

    public String getError() {
        return error;
    }

    public String getStatusCode() {
        return statusCode;
    }
}
