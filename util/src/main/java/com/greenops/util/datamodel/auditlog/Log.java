package com.greenops.util.datamodel.auditlog;

import com.fasterxml.jackson.annotation.JsonSubTypes;
import com.fasterxml.jackson.annotation.JsonTypeInfo;

@JsonTypeInfo(use = JsonTypeInfo.Id.NAME, include = JsonTypeInfo.As.PROPERTY, property = "type")
@JsonSubTypes(
        {
                @JsonSubTypes.Type(value = DeploymentLog.class, name = "deployment"),
                @JsonSubTypes.Type(value = RemediationLog.class, name = "stateremediation")
        }
)
public interface Log {

    public String getPipelineUniqueVersionNumber();

    public int getUniqueVersionInstance();
}
