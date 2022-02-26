package com.greenops.verificationtool.ingest.handling;

import com.greenops.verificationtool.datamodel.status.VerificationStatus;
import com.greenops.verificationtool.datamodel.status.VerificationStatusImpl;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.util.HashMap;

@Component
public class VerificationStatusRegistry {
    private final HashMap<String, VerificationStatus> verificationStatusHashMap;

    @Autowired
    public VerificationStatusRegistry() {
        this.verificationStatusHashMap = new HashMap<>();
    }

    public HashMap<String, VerificationStatus> getVerificationStatusHashMap(){
        return verificationStatusHashMap;
    }

    public VerificationStatus getVerificationStatus(String pipelineName) {
        if (this.verificationStatusHashMap.containsKey(pipelineName)) {
            return this.verificationStatusHashMap.get(pipelineName);
        }
        return null;
    }

    public void putVerificationStatus(String pipelineName) {
        this.verificationStatusHashMap.put(pipelineName, new VerificationStatusImpl());
    }
}
