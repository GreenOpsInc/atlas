package com.greenops.verificationtool.datamodel.verification;

public class Pair {
    private final String step;
    private final Vertex vertex;

    public Pair(String step, Vertex vertex) {
        this.step = step;
        this.vertex = vertex;
    }

    public String getStep() {
        return this.step;
    }

    public Vertex getVertex() {
        return this.vertex;
    }
}