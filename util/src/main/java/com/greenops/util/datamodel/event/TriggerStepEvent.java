package com.greenops.util.datamodel.event;

public class TriggerStepEvent implements Event {

    private String orgName;
    private String teamName;
    private String pipelineName;
    private String stepName;
    private boolean rollback;

    public TriggerStepEvent(String orgName, String teamName, String pipelineName, String stepName, boolean rollback) {
        this.orgName = orgName;
        this.teamName = teamName;
        this.pipelineName = pipelineName;
        this.stepName = stepName;
        this.rollback = rollback;
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

    public boolean isRollback() {
        return rollback;
    }
}