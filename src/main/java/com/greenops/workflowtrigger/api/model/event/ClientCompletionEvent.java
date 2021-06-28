package com.greenops.workflowtrigger.api.model.event;

public class ClientCompletionEvent implements Event {

    private String healthStatus;
    private String orgName;
    private String teamName;
    private String pipelineName;
    private String stepName;
    private String argoName;
    private String operation;
    private String project;
    private String repo;

    public ClientCompletionEvent(String healthStatus, String orgName, String teamName, String pipelineName, String stepName,
                                 String argoName, String operation, String project, String repo) {
        this.healthStatus = healthStatus;
        this.orgName = orgName;
        this.teamName = teamName;
        this.pipelineName = pipelineName;
        this.stepName = stepName;
        this.argoName = argoName;
        this.operation = operation;
        this.project = project;
        this.repo = repo;
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

    @Override
    public String getRepoUrl() {
        return repo;
    }
}
