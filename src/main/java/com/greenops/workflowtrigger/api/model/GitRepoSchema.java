package com.greenops.workflowtrigger.api.model;

public class GitRepoSchema {
    // TODO: The Git repo connection should start with machine accounts (usrname/pass), but then extend to other
    //  methods...this will probably become an interface
    // TODO: Make a test class for GitRepoSchema
    private String gitRepo;
    private String username;
    private String password;

    public GitRepoSchema(String gitRepo, String username, String password) {
        this.gitRepo = gitRepo;
        this.username = username;
        this.password = password;
    }

    public void setGitRepo(String gitRepo) {
        this.gitRepo = gitRepo;
    }

    public void setUsername(String username) {
        this.username = username;
    }

    public void setPassword(String password) {
        this.password = password;
    }

    public String getGitRepo() {
        return gitRepo;
    }

    public String getUsername() {
        return username;
    }

    public String getPassword() {
        return password;
    }
}
