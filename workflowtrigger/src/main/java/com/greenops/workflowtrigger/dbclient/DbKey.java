package com.greenops.workflowtrigger.dbclient;

import org.apache.commons.codec.binary.Base16;
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

    public static String makeDbListOfStepsKey(String orgName, String teamName, String pipelineName) {
        return orgName + "-" + teamName + "-" + pipelineName + "-step";
    }

    public static String makeDbClusterKey(String orgName, String clusterName) {
        return orgName + "-" + clusterName;
    }

    public static String makeSecretName(String orgName, String teamName, String pipelineName) {
        var name = orgName + "-" + teamName + "-" + pipelineName + "-gitcred";
        var base16 = new Base16();
        return base16.encodeAsString(name.getBytes()).toLowerCase();
    }
}
