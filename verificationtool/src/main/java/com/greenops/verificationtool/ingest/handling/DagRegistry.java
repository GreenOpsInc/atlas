package com.greenops.verificationtool.ingest.handling;

import com.greenops.verificationtool.datamodel.verification.DAG;
import com.greenops.verificationtool.ingest.apiclient.workflowtrigger.WorkflowTriggerApi;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.util.HashMap;

@Component
public class DagRegistry {
    private final WorkflowTriggerApi workflowTriggerApi;
    private final HashMap<String, DAG> dagHashMap;

    @Autowired
    public DagRegistry(WorkflowTriggerApi workflowTriggerApi) {
        this.workflowTriggerApi = workflowTriggerApi;
        this.dagHashMap = new HashMap<String, DAG>();
    }

    public void registerDAG(String pipelineName, DAG dagObj) {
        this.dagHashMap.put(pipelineName, dagObj);
    }

    public DAG retriveDagObj(String pipelineName) {
        if (!this.dagHashMap.containsKey(pipelineName)) {
            return null;
        }
        return this.dagHashMap.get(pipelineName);
    }
}