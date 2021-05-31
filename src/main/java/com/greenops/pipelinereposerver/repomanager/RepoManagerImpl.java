package com.greenops.pipelinereposerver.repomanager;

import com.greenops.pipelinereposerver.api.model.git.GitRepoSchema;

import java.io.File;
import java.io.IOException;
import java.util.HashSet;
import java.util.Set;

public class RepoManagerImpl implements RepoManager {

    private final String directory = "tmp";
    private Set<GitRepoSchema> gitRepos;

    public RepoManagerImpl() {
        this.gitRepos = new HashSet<>();
    }

    @Override
    public boolean clone(GitRepoSchema gitRepoSchema) {
        if (gitRepos.stream().anyMatch(gitRepoSchema1 -> gitRepoSchema1.getGitRepo().equals(gitRepoSchema.getGitRepo()))) {
            return true;
        }
        try {
            var command = new CommandBuilder()
                    .gitClone(gitRepoSchema)
                    .build();
            var process = new ProcessBuilder()
                    .command("/bin/bash", "-c", command)
                    .directory(new File(directory))
                    .start();
            int exitCode = process.waitFor();
            if (exitCode == 0) {
                return true;
            } else {
                delete(gitRepoSchema);
                return false;
            }
        } catch (IOException | InterruptedException e) {
            delete(gitRepoSchema);
            return false;
        }
    }

    @Override
    public boolean delete(GitRepoSchema gitRepoSchema) {
        try {
            var command = new CommandBuilder()
                    .deleteFolder(gitRepoSchema)
                    .build();
            var process = new ProcessBuilder()
                    .command("/bin/bash", "-c", command)
                    .directory(new File(directory))
                    .start();
            int exitCode = process.waitFor();
            return exitCode == 0;
        } catch (IOException | InterruptedException e) {
            delete(gitRepoSchema);
            return false;
        }
    }
}
