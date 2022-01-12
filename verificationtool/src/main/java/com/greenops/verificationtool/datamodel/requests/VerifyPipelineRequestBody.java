package com.greenops.verificationtool.datamodel.requests;

import com.greenops.verificationtool.datamodel.rules.RuleData;

import java.util.ArrayList;
import java.util.List;

public class VerifyPipelineRequestBody {
    private final List<RuleData> rules;
    private final String gitRepoUrl;
    private final String pathToRoot;

    public VerifyPipelineRequestBody(String gitRepoUrl, String pathToRoot, List<RuleData> rules) {
        this.rules = rules == null ? new ArrayList<>() : rules;
        this.gitRepoUrl = gitRepoUrl;
        this.pathToRoot = pathToRoot;
    }

    public List<RuleData> getRules() {
        return rules;
    }

    public String getGitRepoUrl() {
        return gitRepoUrl;
    }

    public String getPathToRoot() {
        return pathToRoot;
    }
}
