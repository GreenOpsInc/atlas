package com.greenops.verificationtool.ingest.handling;

import com.greenops.util.datamodel.event.Event;
import org.apache.logging.log4j.util.Strings;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.util.HashMap;
import java.util.List;

@Component
public class EventVisitedRegistry {
    private final HashMap<String, Event> lastVisitedEventHM;

    @Autowired
    public EventVisitedRegistry(){
        lastVisitedEventHM = new HashMap<>();
    }

    private String makeKey(Event event){
        return Strings.join(List.of(event.getOrgName(), event.getTeamName(), event.getPipelineName()), '-');
    }

    public void put(Event event){
        var key = makeKey(event);
        lastVisitedEventHM.put(key, event);
    }

    public Event get(Event event){
        var key = makeKey(event);
        return lastVisitedEventHM.get(key);
    }
}
