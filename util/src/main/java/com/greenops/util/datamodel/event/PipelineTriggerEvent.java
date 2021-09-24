package com.greenops.util.datamodel.event;

import java.util.UUID;

public class PipelineTriggerEvent implements Event {

    public static final String ROOT_STEP_NAME = "ATLAS_ROOT_DATA";

    private String orgName;
    private String teamName;
    private String pipelineName;
    private String uvn;
    private String stepName;

    public PipelineTriggerEvent(String orgName, String teamName, String pipelineName, String pipelineUvn) {
        this.orgName = orgName;
        this.teamName = teamName;
        this.pipelineName = pipelineName;
        this.uvn = pipelineUvn;
        this.stepName = ROOT_STEP_NAME;
    }

    public PipelineTriggerEvent(String orgName, String teamName, String pipelineName) {
        this(orgName, teamName, pipelineName, UUID.randomUUID().toString());
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
        return uvn;
    }

    @Override
    public String getStepName() {
        return stepName;
    }
}
