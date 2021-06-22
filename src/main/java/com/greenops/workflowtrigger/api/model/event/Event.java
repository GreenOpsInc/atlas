package com.greenops.workflowtrigger.api.model.event;

import com.fasterxml.jackson.annotation.JsonSubTypes;
import com.fasterxml.jackson.annotation.JsonTypeInfo;

@JsonTypeInfo(use = JsonTypeInfo.Id.NAME, include = JsonTypeInfo.As.PROPERTY, property = "type")
@JsonSubTypes(
        {
                @JsonSubTypes.Type(value = ClientCompletionEvent.class, name = "clientcompletion")
        }
)
public interface Event {
    String getPipelineName();
    String getStepName();
    String getRepoUrl();
}
