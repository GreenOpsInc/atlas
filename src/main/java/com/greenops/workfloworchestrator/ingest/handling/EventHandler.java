package com.greenops.workfloworchestrator.ingest.handling;

import com.greenops.workfloworchestrator.datamodel.event.Event;

public interface EventHandler {
    void handleEvent(Event event);
}
