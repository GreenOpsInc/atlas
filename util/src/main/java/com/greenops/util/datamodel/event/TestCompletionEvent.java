package com.greenops.util.datamodel.event;

import org.apache.logging.log4j.util.Strings;

import java.util.List;

public class TestCompletionEvent implements Event {

    private boolean successful;
    private String orgName;
    private String teamName;
    private String pipelineName;
    private String pipelineUvn;
    private String stepName;
    private String log;
    private String testName;
    private int testNumber;
    private int deliveryAttempt;

    public TestCompletionEvent(boolean successful, String orgName, String teamName, String pipelineName, String pipelineUvn, String stepName,
                                 String log, String testName, int testNumber) {
        this.successful = successful;
        this.orgName = orgName;
        this.teamName = teamName;
        this.pipelineName = pipelineName;
        this.pipelineUvn = pipelineUvn;
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
        return pipelineUvn;
    }

    @Override
    public String getStepName() {
        return stepName;
    }

    //Uvn is unique to a pipeline, no two tests will have the same test number
    @Override
    public String getMQKey() {
        return Strings.join(List.of(pipelineUvn, stepName, testNumber, getDeliveryAttempt()), '-');
    }

    @Override
    public void setDeliveryAttempt(int attempt) {
        this.deliveryAttempt = attempt;
    }

    @Override
    public int getDeliveryAttempt() {
        return deliveryAttempt;
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
