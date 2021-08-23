package com.greenops.workfloworchestrator.datamodel.mixin.requests;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.greenops.workfloworchestrator.datamodel.requests.ResourceGvk;

import java.util.List;

public abstract class ResourcesGvkRequestMixin {

    @JsonProperty(value = "resourceGvkList")
    List<ResourceGvk> resourceGvkList;

    @JsonIgnore
    public ResourcesGvkRequestMixin(@JsonProperty(value = "resourceGvkList") List<ResourceGvk> resourceGvkList) {
    }

    @JsonIgnore
    abstract void addResource(String resourceName, String resourceNamespace, String group, String version, String kind);
}
