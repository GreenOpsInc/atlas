package com.greenops.util.datamodel.event;

public class ApplicationInfraTriggerEvent implements Event {

    private String orgName;
    private String teamName;
    private String pipelineName;
    private String stepName;

    public ApplicationInfraTriggerEvent(String orgName, String teamName, String pipelineName, String stepName) {
        this.orgName = orgName;
        this.teamName = teamName;
        this.pipelineName = pipelineName;
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
    public String getStepName() {
        return stepName;
    }
}
