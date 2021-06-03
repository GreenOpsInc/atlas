package com.greenops.workflowtrigger.dbclient;

public class DbKey {

    //TODO: Going forward, there will be different types of keys. Team keys will be <org_name>-<team_name>,
    //the list of teams key would look like <org_name>-teams, etc. We should also be using regex as the number of key
    //types grow in the future.

    public static String makeDbTeamKey(String orgName, String teamName) {
        return orgName + "-" + teamName;
    }
}
