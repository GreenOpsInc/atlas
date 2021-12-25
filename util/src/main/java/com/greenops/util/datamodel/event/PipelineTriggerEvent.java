package com.greenops.util.datamodel.event;

import java.util.UUID;

public class PipelineTriggerEvent implements Event {

    public static final String ROOT_STEP_NAME = "ATLAS_ROOT_DATA";
    public static final String ROOT_COMMIT = "ROOT_COMMIT";
    public static final String PIPELINE_TRIGGER_EVENT_CLASS_NAME = "com.greenops.util.datamodel.event.PipelineTriggerEvent";

    private String orgName;
    private String teamName;
    private String pipelineName;
    private String uvn;
    private String stepName;
    private String revisionHash;

    public PipelineTriggerEvent(String orgName, String teamName, String pipelineName, String pipelineUvn, String stepName, String revisionHash) {
        this.orgName = orgName;
        this.teamName = teamName;
        this.pipelineName = pipelineName;
        this.uvn = pipelineUvn;
        this.stepName = stepName;
        this.revisionHash = revisionHash;
    }

    public PipelineTriggerEvent(String orgName, String teamName, String pipelineName) {
        this(orgName, teamName, pipelineName, UUID.randomUUID().toString(), ROOT_STEP_NAME, ROOT_COMMIT);
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

    public String getRevisionHash() {
        return revisionHash;
    }
}
