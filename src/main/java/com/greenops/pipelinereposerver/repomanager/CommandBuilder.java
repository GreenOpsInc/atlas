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

    CommandBuilder deleteFolder(GitRepoSchema gitRepoSchema) {
        var newCommand = new ArrayList<String>();
        newCommand.add("rm -rf");
        var splitLink = gitRepoSchema.getGitRepo().split("/");
        int idx = splitLink.length - 1;
        while (idx >= 0) {
            if (splitLink[idx].equals("")) {
                idx--;
            } else {
                break;
            }
        }
        newCommand.add(splitLink[idx].replace(".git", ""));
        commands.add(String.join(" ", newCommand));
        return this;
    }

    String build() {
        return String.join("; ", commands);
    }
}
