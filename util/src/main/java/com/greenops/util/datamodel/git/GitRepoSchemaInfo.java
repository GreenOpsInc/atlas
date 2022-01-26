package com.greenops.util.datamodel.git;

public class GitRepoSchemaInfo {

    private String gitRepo;
    private String pathToRoot;

    public GitRepoSchemaInfo(String gitRepo, String pathToRoot) {
        this.gitRepo = gitRepo;
        this.pathToRoot = pathToRoot;
    }

    public void setGitRepo(String gitRepo) {
        this.gitRepo = gitRepo;
    }

    public void setPathToRoot(String pathToRoot) {
        this.pathToRoot = pathToRoot;
    }

    public String getGitRepo() {
        return gitRepo;
    }

    public String getPathToRoot() {
        return pathToRoot;
    }

    @Override
    public boolean equals(Object o) {
        if (o instanceof GitRepoSchemaInfo) {
            return ((GitRepoSchemaInfo) o).gitRepo.equals(this.gitRepo)
                    && ((GitRepoSchemaInfo) o).pathToRoot.equals(this.pathToRoot);
        }

        return false;
    }
}
