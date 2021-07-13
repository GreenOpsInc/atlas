package com.greenops.pipelinereposerver.repomanager;

import com.greenops.pipelinereposerver.api.model.git.GitRepoSchema;

import java.util.ArrayList;
import java.util.List;

public class CommandBuilder {
    private static final String GIT_SUFFIX = ".git";
    private List<String> commands;

    CommandBuilder() {
        commands = new ArrayList<>();
    }

    CommandBuilder mkdir(String directory) {
        commands.add("mkdir " + directory);
        return this;
    }

    CommandBuilder cd(String directory) {
        commands.add("cd " + directory);
        return this;
    }

    CommandBuilder gitClone(GitRepoSchema gitRepoSchema) {
        var newCommand = new ArrayList<String>();
        newCommand.add("git clone");
        newCommand.add(gitRepoSchema.getGitCred().convertGitCredToString(gitRepoSchema.getGitRepo()));
        commands.add(String.join(" ", newCommand));
        return this;
    }

    CommandBuilder gitPull(GitRepoSchema gitRepoSchema) {
        var newCommand = new ArrayList<String>();
        newCommand.add("git pull");
        newCommand.add(gitRepoSchema.getGitCred().convertGitCredToString(gitRepoSchema.getGitRepo()));
        commands.add(String.join(" ", newCommand));
        return this;
    }

    //Can only be used with an existing repo. Remotes will not work.
    CommandBuilder gitLog(int logCount, boolean justCommits) {
        if (justCommits) {
            commands.add("git log -n " + logCount + " --pretty=format:\"%H\"");
        } else {
            commands.add("git log -n " + logCount);
        }
        return this;
    }

    CommandBuilder deleteFolder(GitRepoSchema gitRepoSchema) {
        var newCommand = new ArrayList<String>();
        newCommand.add("rm -rf");
        newCommand.add(getFolderName(gitRepoSchema.getGitRepo()));
        commands.add(String.join(" ", newCommand));
        return this;
    }

    String build() {
        return String.join("; ", commands);
    }

    public static String getFolderName(String gitRepo) {
        var splitLink = gitRepo.split("/");
        int idx = splitLink.length - 1;
        while (idx >= 0) {
            if (splitLink[idx].equals("")) {
                idx--;
            } else {
                break;
            }
        }
        return splitLink[idx].replace(".git", "");
    }
}
