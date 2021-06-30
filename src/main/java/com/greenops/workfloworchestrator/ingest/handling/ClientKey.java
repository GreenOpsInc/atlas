package com.greenops.workfloworchestrator.ingest.handling;

import org.apache.logging.log4j.util.Strings;

import java.util.List;
import java.util.UUID;

public class ClientKey {
    static String makeTestKey(int testNumber) {
        return Strings.join(List.of(testNumber, UUID.randomUUID()), '-');
    }
}
