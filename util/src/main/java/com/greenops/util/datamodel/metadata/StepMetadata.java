package com.greenops.util.datamodel.metadata;

import com.greenops.util.datamodel.git.ArgoRepoSchema;

public class StepMetadata {

    private ArgoRepoSchema argoRepoSchema;

    public StepMetadata(ArgoRepoSchema argoRepoSchema) {
        this.argoRepoSchema = argoRepoSchema;
    }

    public ArgoRepoSchema getArgoRepoSchema() {
        return argoRepoSchema;
    }

    public void setArgoRepoSchema(ArgoRepoSchema argoRepoSchema) {
        this.argoRepoSchema = argoRepoSchema;
    }
}
