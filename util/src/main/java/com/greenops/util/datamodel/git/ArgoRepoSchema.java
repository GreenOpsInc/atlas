package com.greenops.util.datamodel.git;

public class ArgoRepoSchema {

    private String repoUrl;
    private String targetRevision;
    private String path;

    public ArgoRepoSchema(String repoUrl, String targetRevision, String path) {
        this.repoUrl = repoUrl;
        this.targetRevision = targetRevision.isEmpty() ? "main" : targetRevision;
        this.path = path.isEmpty() ? "/" : path;
    }

    public String getRepoUrl() {
        return repoUrl;
    }

    public String getTargetRevision() {
        return targetRevision;
    }

    public String getPath() {
        return path;
    }

    @Override
    public boolean equals(Object otherArgoRepoSchema) {
        assert otherArgoRepoSchema instanceof ArgoRepoSchema;
        return repoUrl.equals(((ArgoRepoSchema) otherArgoRepoSchema).repoUrl)
                && targetRevision.equals(((ArgoRepoSchema) otherArgoRepoSchema).targetRevision)
                && path.equals(((ArgoRepoSchema) otherArgoRepoSchema).path);
    }
}
