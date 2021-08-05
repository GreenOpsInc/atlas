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

    public static String makeDbListOfTeamsKey(String orgName) {
        return orgName + "-teams";
    }
}
