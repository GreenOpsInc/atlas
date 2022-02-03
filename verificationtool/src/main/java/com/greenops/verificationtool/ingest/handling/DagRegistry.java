package com.greenops.verificationtool.ingest.handling;

import com.greenops.util.datamodel.event.Event;
import com.greenops.util.datamodel.event.PipelineCompletionEvent;
import com.greenops.verificationtool.datamodel.verification.DAG;
import com.greenops.verificationtool.ingest.apiclient.workflowtrigger.WorkflowTriggerApi;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.util.HashMap;

@Component
public class DagRegistry {
    private final String COMPLETED = "completed";
    private final String FAILED = "failed";
    private final String PROGRESS = "progress";
    private final String NOT_FOUND = "NOT_FOUND";
    private final WorkflowTriggerApi workflowTriggerApi;
    private final HashMap<String, DAG> dagHashMap;
    private final HashMap<String, String> verificationStatus;

    @Autowired
    public DagRegistry(WorkflowTriggerApi workflowTriggerApi) {
        this.workflowTriggerApi = workflowTriggerApi;
        this.dagHashMap = new HashMap<String, DAG>();
        this.verificationStatus = new HashMap<String, String>();
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

    public HashMap<String, String> getPipelineStatus() {
        return this.verificationStatus;
    }

    public String getSinglePipelineStatus(String pipelineName) {
        if (this.verificationStatus.get(pipelineName) == null) {
            return this.NOT_FOUND;
        }
        return this.verificationStatus.get(pipelineName);
    }

    public void markPipelineProgress(Event event) {
        if (event instanceof PipelineCompletionEvent) {
            this.verificationStatus.put(event.getPipelineName(), this.COMPLETED);
        } else {
            this.verificationStatus.put(event.getPipelineName(), this.PROGRESS);
        }
    }
}