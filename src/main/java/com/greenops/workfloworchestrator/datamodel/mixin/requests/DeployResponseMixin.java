package com.greenops.workfloworchestrator.datamodel.mixin.requests;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class DeployResponseMixin {

    @JsonProperty(value = "Success")
    boolean success;
    @JsonProperty(value = "ResourceName")
    String resourceName;
    @JsonProperty(value = "AppNamespace")
    String applicationNamespace;
    @JsonProperty(value = "RevisionHash")
    String revisionHash;

    @JsonCreator
    public DeployResponseMixin(@JsonProperty(value = "Success") boolean success,
                               @JsonProperty(value = "ResourceName") String resourceName,
                               @JsonProperty(value = "AppNamespace") String applicationNamespace,
                               @JsonProperty(value = "RevisionHash") String revisionHash) {
    }
}
