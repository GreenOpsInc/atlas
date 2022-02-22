package com.greenops.verificationtool.ingest.handling;

import com.greenops.util.datamodel.event.Event;

public interface EventHandler {
    void handleEvent(Event event) throws InterruptedException;
}
