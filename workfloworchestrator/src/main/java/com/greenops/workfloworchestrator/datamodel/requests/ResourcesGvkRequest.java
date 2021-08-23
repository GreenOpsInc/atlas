package com.greenops.workfloworchestrator.datamodel.requests;

import java.util.ArrayList;
import java.util.List;

public class ResourcesGvkRequest {

    private List<ResourceGvk> resourceGvkList;

    public ResourcesGvkRequest() {
        this.resourceGvkList = new ArrayList<>();
    }

    public ResourcesGvkRequest(List<ResourceGvk> resourceGvkList) {
        this.resourceGvkList = resourceGvkList;
    }

    public void addResource(String resourceName, String resourceNamespace, String group, String version, String kind) {
        this.resourceGvkList.add(new ResourceGvk(resourceName, resourceNamespace, group, version, kind));
    }
}
