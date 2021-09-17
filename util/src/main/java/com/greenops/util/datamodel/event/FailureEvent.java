package com.greenops.util.datamodel.event;

import com.greenops.util.datamodel.request.DeployResponse;

public class FailureEvent implements Event {

    private String orgName;
    private String teamName;
    private String pipelineName;
    private String stepName;
    private DeployResponse deployResponse;
    private String statusCode;
    private String error;

    public FailureEvent(String orgName, String teamName, String pipelineName, String stepName, DeployResponse deployResponse, String statusCode, String error) {
        this.orgName = orgName;
        this.teamName = teamName;
        this.pipelineName = pipelineName;
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
    public String getStepName() {
        return stepName;
    }

    public DeployResponse getDeployResponse() {
        return deployResponse;
    }

    public String getError() {
        return error;
    }
}
