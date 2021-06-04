package com.greenops.workflowtrigger.dbclient;

public class DbKey {

    public static String makeDbTeamKey(String orgName, String teamName) {
        return orgName + "-" + teamName;
    }

    public static String makeDbListOfTeamsKey(String orgName) {
        return orgName + "-teams";
    }
}
