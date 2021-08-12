package com.greenops.util.datamodel.mixin.metadata;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.greenops.util.datamodel.git.ArgoRepoSchema;

public abstract class StepMetadataMixin {

    @JsonProperty("argoRepoSchema")
    ArgoRepoSchema argoRepoSchema;

    @JsonCreator
    public StepMetadataMixin(@JsonProperty("argoRepoSchema") ArgoRepoSchema argoRepoSchema) {
    }
}
