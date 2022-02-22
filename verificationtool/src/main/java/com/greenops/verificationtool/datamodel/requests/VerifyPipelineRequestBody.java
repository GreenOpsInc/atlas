package com.greenops.verificationtool.datamodel.requests;

import com.greenops.verificationtool.datamodel.rules.RuleData;

import java.util.ArrayList;
import java.util.List;

public class VerifyPipelineRequestBody {
    private final String gitRepoUrl;
    private final String pathToRoot;
    private final String teamName;

    public VerifyPipelineRequestBody(String gitRepoUrl, String pathToRoot, String teamName) {
        this.gitRepoUrl = gitRepoUrl;
        this.pathToRoot = pathToRoot;
        this.teamName = teamName;
    }

    public String getGitRepoUrl() {
        return gitRepoUrl;
    }

    public String getPathToRoot() {
        return pathToRoot;
    }

    public String getTeamName() {
        return teamName;
    }
}
