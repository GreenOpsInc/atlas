package com.greenops.workfloworchestrator.ingest.handling;

import org.apache.logging.log4j.util.Strings;

import java.util.List;
import java.util.UUID;

public class ClientKey {
    public static String makeTestKey(int testNumber) {
        return Strings.join(List.of(testNumber, UUID.randomUUID()), '-');
    }

    public static int getTestNumberFromTestKey(String testKey) {
        var splitKey = testKey.split("-");
        if (splitKey.length > 0) {
            return Integer.parseInt(splitKey[0]);
        }
        return -1;
    }
}
