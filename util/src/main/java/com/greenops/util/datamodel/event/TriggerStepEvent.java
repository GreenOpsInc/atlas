package com.greenops.util.datamodel.event;

public class TriggerStepEvent implements Event {

    private String orgName;
    private String teamName;
    private String pipelineName;
    private String stepName;
    private String argoRevisionHash;
    private String pipelineUvn;
    private boolean rollback;

    public TriggerStepEvent(String orgName, String teamName, String pipelineName, String stepName, String argoRevisionHash, String pipelineUvn, boolean rollback) {
        this.orgName = orgName;
        this.teamName = teamName;
        this.pipelineName = pipelineName;
        this.stepName = stepName;
        this.argoRevisionHash = argoRevisionHash;
        this.pipelineUvn = pipelineUvn;
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

    public String getArgoRevisionHash() {
        return argoRevisionHash;
    }

    public String getPipelineUvn() {
        return pipelineUvn;
    }

    public boolean isRollback() {
        return rollback;
    }
}
