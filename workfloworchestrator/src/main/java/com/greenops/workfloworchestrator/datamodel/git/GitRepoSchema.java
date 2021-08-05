package com.greenops.workfloworchestrator.datamodel.git;

public class GitRepoSchema {
    // TODO: The Git repo connection should start with machine accounts (usrname/pass), but then extend to other
    //  methods...this will probably become an interface
    // TODO: Make a test class for GitRepoSchema

    private static String gitRepoName = "Git Repo";
    private static String rootName = "Root";
    private static String gitCredName = "Git Cred";

    private String gitRepo;
    private String pathToRoot;
    private GitCred gitCred;

    public GitRepoSchema(String gitRepo, String pathToRoot, GitCred gitCred) {
        this.gitRepo = gitRepo;
        this.pathToRoot = pathToRoot;
        this.gitCred = gitCred;
    }

    public void setGitRepo(String gitRepo) {
        this.gitRepo = gitRepo;
    }

    public void setPathToRoot(String pathToRoot) {
        this.pathToRoot = pathToRoot;
    }

    public void setGitCred(GitCred gitCred) {
        this.gitCred = gitCred;
    }

    public String getGitRepo() {
        return gitRepo;
    }

    public String getPathToRoot() {
        return pathToRoot;
    }
}
