package com.greenops.workflowtrigger.api.model.git;

import com.fasterxml.jackson.annotation.JsonSubTypes;
import com.fasterxml.jackson.annotation.JsonTypeInfo;

@JsonTypeInfo(use = JsonTypeInfo.Id.NAME, include = JsonTypeInfo.As.PROPERTY, property = "type")
@JsonSubTypes(
        {
                @JsonSubTypes.Type(value = GitCredMachineUser.class, name = "machineuser"),
                @JsonSubTypes.Type(value = GitCredOpen.class, name = "open"),
                @JsonSubTypes.Type(value = GitCredToken.class, name = "oauth")
        }
)
public interface GitCred {}
