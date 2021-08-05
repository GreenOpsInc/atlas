package com.greenops.pipelinereposerver.dbclient;

import org.apache.commons.codec.binary.Base16;

public class DbKey {

    public static String makeDbTeamKey(String orgName, String teamName) {
        return orgName + "-" + teamName;
    }

    public static String makeDbListOfTeamsKey(String orgName) {
        return orgName + "-teams";
    }

    public static String makeSecretName(String orgName, String teamName, String pipelineName) {
        var name = orgName + "-" + teamName + "-" + pipelineName + "-gitcred";
        var base16 = new Base16();
        return base16.encodeAsString(name.getBytes()).toLowerCase();
    }
}
