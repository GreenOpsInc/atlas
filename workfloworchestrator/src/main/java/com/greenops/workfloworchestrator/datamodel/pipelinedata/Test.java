package com.greenops.workfloworchestrator.datamodel.pipelinedata;

import com.fasterxml.jackson.annotation.JsonSubTypes;
import com.fasterxml.jackson.annotation.JsonTypeInfo;

import java.util.Map;

@JsonTypeInfo(use = JsonTypeInfo.Id.NAME, include = JsonTypeInfo.As.PROPERTY, property = "type")
@JsonSubTypes(
        {
                @JsonSubTypes.Type(value = InjectScriptTest.class, name = "inject"),
                @JsonSubTypes.Type(value = CustomJobTest.class, name = "custom")
        }
)
public interface Test {
    public String getPath();
    public boolean shouldExecuteBefore();
    public Map<String, String> getVariables();
    //The expectation is that getPayload will return either a String or a KubernetesCreationRequest
    public Object getPayload(int testNumber, String config, String teamName, String pipelineName, String testConfig);
}
