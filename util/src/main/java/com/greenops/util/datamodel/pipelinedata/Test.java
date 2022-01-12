package com.greenops.util.datamodel.pipelinedata;

import com.fasterxml.jackson.annotation.JsonSubTypes;
import com.fasterxml.jackson.annotation.JsonTypeInfo;

import java.util.Map;

@JsonTypeInfo(use = JsonTypeInfo.Id.NAME, include = JsonTypeInfo.As.PROPERTY, property = "type")
@JsonSubTypes(
        {
                @JsonSubTypes.Type(value = InjectScriptTest.class, name = Test.INJECT_TASK),
                @JsonSubTypes.Type(value = CustomJobTest.class, name = Test.CUSTOM_TASK),
        }
)
public interface Test {
    String INJECT_TASK = "inject";
    String CUSTOM_TASK = "custom";

    String getPath();

    boolean shouldExecuteBefore();

    Map<String, String> getVariables();

    //The expectation is that getPayload will return either a String or a KubernetesCreationRequest
    Object getPayload(int testNumber, String testConfig);

    String getWatchKey();
}
