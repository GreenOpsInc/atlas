package com.greenops.workfloworchestrator.ingest.handling;

import org.apache.logging.log4j.util.Strings;

import java.util.List;

public class ClientKey {
    static String makeTestKey(String teamName, String pipelineName, String stepName, String testName) {
        return Strings.join(List.of(teamName, pipelineName, "stepName", testName), '-').toLowerCase();
    }
}
