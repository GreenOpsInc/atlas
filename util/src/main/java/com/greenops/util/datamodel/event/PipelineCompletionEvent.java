package com.greenops.util.datamodel.event;

// This Event is only used by verificationtool
public class PipelineCompletionEvent implements Event {
    public static final String ROOT_STEP_NAME = "ATLAS_ROOT_DATA";
    public static final String ROOT_COMMIT = "ROOT_COMMIT";
    public static final String PIPELINE_TRIGGER_EVENT_CLASS_NAME = "com.greenops.util.datamodel.event.PipelineCompletionEvent";

    private final String orgName;
    private final String teamName;
    private final String pipelineName;
    private final String uvn;
    private final String stepName;
    private final String revisionHash;

    public PipelineCompletionEvent(String orgName, String teamName, String pipelineName, String pipelineUvn) {
        this.orgName = orgName;
        this.teamName = teamName;
        this.pipelineName = pipelineName;
        this.uvn = pipelineUvn;
        this.stepName = ROOT_STEP_NAME;
        this.revisionHash = ROOT_COMMIT;
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
