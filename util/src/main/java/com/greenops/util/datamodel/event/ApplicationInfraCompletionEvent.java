package com.greenops.util.datamodel.event;

public class ApplicationInfraCompletionEvent implements Event {

    private String orgName;
    private String teamName;
    private String pipelineName;
    private String stepName;
    private boolean success;

    public ApplicationInfraCompletionEvent(String orgName, String teamName, String pipelineName, String stepName, boolean success) {
        this.orgName = orgName;
        this.teamName = teamName;
        this.pipelineName = pipelineName;
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
    public String getStepName() {
        return stepName;
    }

    public boolean isSuccess() {
        return success;
    }
}