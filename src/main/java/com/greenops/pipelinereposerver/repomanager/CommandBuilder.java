package com.greenops.pipelinereposerver.repomanager;

import com.greenops.pipelinereposerver.api.model.git.GitRepoSchema;

import java.util.ArrayList;
import java.util.List;

public class CommandBuilder {
    private List<String> commands;

    CommandBuilder() {
        commands = new ArrayList<>();
    }

    CommandBuilder gitClone(GitRepoSchema gitRepoSchema) {
        var newCommand = new ArrayList<String>();
        newCommand.add("git clone");
        newCommand.add(gitRepoSchema.getGitCred().convertGitCredToString(gitRepoSchema.getGitRepo()));
        commands.add(String.join(" ", newCommand));
        return this;
    }

    String build() {
        return String.join("; ", commands);
    }
}
