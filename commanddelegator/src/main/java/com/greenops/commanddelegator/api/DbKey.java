package com.greenops.commanddelegator.api;

import org.apache.logging.log4j.util.Strings;

import java.util.List;

public class DbKey {

    public static String makeClientRequestQueueKey(String orgName, String clusterName) {
        return Strings.join(List.of(orgName, clusterName, "events"), '-');
    }

    public static String makeDbClusterKey(String orgName, String clusterName) {
        return orgName + "-" + clusterName;
    }
}
