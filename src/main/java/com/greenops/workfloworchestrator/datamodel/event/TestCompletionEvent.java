package com.greenops.workfloworchestrator.datamodel.event;

public class TestCompletionEvent implements Event {

    private boolean successful;
    private String orgName;
    private String teamName;
    private String pipelineName;
    private String stepName;
    private String log;
    private String testName;

    public TestCompletionEvent(boolean successful, String orgName, String teamName, String pipelineName, String stepName,
                                 String log, String testName) {
        this.successful = successful;
        this.orgName = orgName;
        this.teamName = teamName;
        this.pipelineName = pipelineName;
        this.stepName = stepName;
        this.log = log;
        this.testName = testName;
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

    public boolean getSuccessful() {
        return successful;
    }

    public String getTestName() {
        return testName;
    }
}
