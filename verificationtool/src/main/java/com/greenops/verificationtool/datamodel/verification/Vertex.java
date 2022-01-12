package com.greenops.verificationtool.datamodel.verification;

public class Vertex {
    private final String eventType;
    private final String pipelineName;
    private final String stepName;
    private final Integer testNumber;

    public Vertex(String eventType, String pipelineName, String stepName) {
        this.eventType = eventType;
        this.pipelineName = pipelineName;
        this.stepName = stepName;
        this.testNumber = -1;
    }

    public Vertex(String eventType, String pipelineName, String stepName, Integer testNumber) {
        this.eventType = eventType;
        this.pipelineName = pipelineName;
        this.stepName = stepName;
        this.testNumber = testNumber;
    }

    public Integer getTestNumber() {
        return this.testNumber;
    }

    public String getEventType() {
        return this.eventType;
    }

    public String getPipelineName() {
        return this.pipelineName;
    }

    public String getStepName() {
        return stepName;
    }

    public String toString() {
        return this.eventType + " " + this.pipelineName + " " + this.stepName + " " + this.testNumber;
    }

    @Override
    public int hashCode() {
        return eventType.hashCode() * pipelineName.hashCode() * stepName.hashCode() * testNumber.hashCode();
    }

    @Override
    public boolean equals(Object obj) {
        return obj instanceof Vertex
                && this.eventType.equals(((Vertex) obj).eventType)
                && this.pipelineName.equals(((Vertex) obj).pipelineName)
                && this.stepName.equals(((Vertex) obj).stepName)
                && this.testNumber.equals(((Vertex) obj).testNumber);
    }
}