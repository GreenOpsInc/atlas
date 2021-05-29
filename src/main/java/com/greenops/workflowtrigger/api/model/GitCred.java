package com.greenops.workflowtrigger.api.model;

import com.fasterxml.jackson.annotation.JsonSubTypes;
import com.fasterxml.jackson.annotation.JsonTypeInfo;
import org.json.JSONObject;

@JsonTypeInfo(use = JsonTypeInfo.Id.NAME, include = JsonTypeInfo.As.PROPERTY, property = "type")
@JsonSubTypes(
        {
                @JsonSubTypes.Type(value = GitCredMachineUser.class, name = "machineuser"),
                @JsonSubTypes.Type(value = GitCredOpen.class, name = "open")
        }
)
public interface GitCred {
    JSONObject convertToJson();
}
