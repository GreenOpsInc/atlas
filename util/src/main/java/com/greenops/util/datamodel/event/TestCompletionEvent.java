package com.greenops.util.datamodel.event;

public class TestCompletionEvent implements Event {

    private boolean successful;
    private String orgName;
    private String teamName;
    private String pipelineName;
    private String uvn;
    private String stepName;
    private String log;
    private String testName;
    private int testNumber;

    public TestCompletionEvent(boolean successful, String orgName, String teamName, String pipelineName, String uvn, String stepName,
                                 String log, String testName, int testNumber) {
        this.successful = successful;
        this.orgName = orgName;
        this.teamName = teamName;
        this.pipelineName = pipelineName;
        this.uvn = uvn;
        this.stepName = stepName;
        this.log = log;
        this.testName = testName;
        this.testNumber = testNumber;
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

    public boolean getSuccessful() {
        return successful;
    }

    public String getTestName() {
        return testName;
    }

    public String getLog() {
        return log;
    }

    public int getTestNumber() {
        return testNumber;
    }
}
