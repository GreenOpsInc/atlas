package com.greenops.util.datamodel.event;

import com.fasterxml.jackson.annotation.JsonSubTypes;
import com.fasterxml.jackson.annotation.JsonTypeInfo;

@JsonTypeInfo(use = JsonTypeInfo.Id.NAME, include = JsonTypeInfo.As.PROPERTY, property = "type")
@JsonSubTypes(
        {
                @JsonSubTypes.Type(value = ClientCompletionEvent.class, name = "clientcompletion"),
                @JsonSubTypes.Type(value = TestCompletionEvent.class, name = "testcompletion"),
                @JsonSubTypes.Type(value = ApplicationInfraTriggerEvent.class, name = "appinfratrigger"),
                @JsonSubTypes.Type(value = ApplicationInfraCompletionEvent.class, name = "appinfracompletion"),
                @JsonSubTypes.Type(value = TriggerStepEvent.class, name = "triggerstep")
        }
)
public interface Event {
    String getOrgName();
    String getTeamName();
    String getPipelineName();
    String getStepName();
}
