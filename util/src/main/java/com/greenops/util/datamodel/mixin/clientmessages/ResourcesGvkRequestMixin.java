package com.greenops.util.datamodel.mixin.clientmessages;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.greenops.util.datamodel.clientmessages.ResourceGvk;

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
