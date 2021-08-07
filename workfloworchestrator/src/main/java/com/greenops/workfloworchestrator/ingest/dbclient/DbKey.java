package com.greenops.workfloworchestrator.ingest.dbclient;

import org.apache.logging.log4j.util.Strings;

import java.util.List;

public class DbKey {

    public static String makeDbTeamKey(String orgName, String teamName) {
        return orgName + "-" + teamName;
    }

    public static String makeDbStepKey(String orgName, String teamName, String pipelineName, String stepName) {
        return Strings.join(List.of(orgName, teamName, pipelineName, stepName), '-');
    }

    public static String makeClientRequestQueueKey(String orgName, String teamName, String clusterName) {
        return Strings.join(List.of(orgName, teamName, clusterName, "events"), '-');
    }

    public static String makeDbListOfTeamsKey(String orgName) {
        return orgName + "-teams";
    }

    public static String makeDbListOfStepsKey(String orgName, String teamName, String pipelineName) {
        return orgName + "-" + teamName + "-" + pipelineName + "-step";
    }
}
