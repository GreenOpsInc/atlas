package com.greenops.util.datamodel.pipelinedata;

import com.fasterxml.jackson.annotation.JsonSubTypes;
import com.fasterxml.jackson.annotation.JsonTypeInfo;

import java.util.Map;

@JsonTypeInfo(use = JsonTypeInfo.Id.NAME, include = JsonTypeInfo.As.PROPERTY, property = "type")
@JsonSubTypes(
        {
                @JsonSubTypes.Type(value = InjectScriptTest.class, name = Test.INJECT_TASK),
                @JsonSubTypes.Type(value = CustomJobTest.class, name = Test.CUSTOM_TASK),
                @JsonSubTypes.Type(value = ArgoWorkflowTask.class, name = Test.ARGO_WORKFLOW_TASK)
        }
)
public interface Test {
    static final String INJECT_TASK = "inject";
    static final String CUSTOM_TASK = "custom";
    static final String ARGO_WORKFLOW_TASK = "ArgoWorkflowTask";

    public String getPath();
    public boolean shouldExecuteBefore();
    public Map<String, String> getVariables();
    //The expectation is that getPayload will return either a String or a KubernetesCreationRequest
    public Object getPayload(int testNumber, String testConfig);
    public String getWatchKey();
}
