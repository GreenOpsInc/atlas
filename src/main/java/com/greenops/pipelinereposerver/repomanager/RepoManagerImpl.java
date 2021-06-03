package com.greenops.pipelinereposerver.repomanager;

import com.greenops.pipelinereposerver.api.model.git.GitRepoSchema;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

import java.io.File;
import java.io.IOException;
import java.util.HashSet;
import java.util.Set;

@Slf4j
@Component
public class RepoManagerImpl implements RepoManager {

    private final String orgName = "temporary"; //TODO: Needs to be updated when we decide how to configure organization
    private final String directory = "tmp";
    private Set<GitRepoSchema> gitRepos;

    public RepoManagerImpl() {
        this.gitRepos = new HashSet<>();
        if (!setupGitCli()) {
            throw new RuntimeException("Git could not be installed. Something may be wrong with the configuration.");
        }
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
                    .directory(new File(orgName + "/" + directory))
                    .start();
            int exitCode = process.waitFor();
            if (exitCode == 0) {
                log.info("Cloning repo {} was successful.", gitRepoSchema.getGitRepo());
                return true;
            } else {
                log.info("Cloning repo {} was not successful. Cleaning up...", gitRepoSchema.getGitRepo());
                delete(gitRepoSchema);
                return false;
            }
        } catch (IOException | InterruptedException e) {
            log.error("An error was thrown when attempting to clone the repo {}", gitRepoSchema.getGitRepo(), e);
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
                    .directory(new File(orgName + "/" + directory))
                    .start();
            int exitCode = process.waitFor();
            log.info("Deletion of repo {} finished with exit code {}", gitRepoSchema.getGitRepo(), exitCode);
            return exitCode == 0;
        } catch (IOException | InterruptedException e) {
            log.error("An error was thrown when attempting to delete the repo {}", gitRepoSchema.getGitRepo(), e);
            return false;
        }
    }

    public boolean setupGitCli() {
        //TODO: This method is absolutely temporary. In no way should this be the de-facto way of installing tools going forward.
        //Jib is a little weird in the sense that it doesn't allow traditional docker builds and doesn't allow RUN commands.
        //Jib can take a custom base image, which we will have to configure going forwards.
        //This temporality also applies for doing mkdir directory & cd directory. We should be adding the creation of the directory
        //to the container itself, and also using the .directory(...) ProcessBuilder command that is commented out for future usage.
        try {
            var command = new CommandBuilder()
                    .mkdir(orgName)
                    .cd(orgName)
                    .mkdir(directory)
                    .cd(directory)
                    .build();
            var process = new ProcessBuilder()
                    .command("/bin/bash", "-c", "apt-get update && apt-get install -y git; ls tmp;" + command)
                    .start();
            int exitCode = process.waitFor();
            return exitCode == 0;
        } catch (IOException | InterruptedException e) {
            return false;
        }
    }
}
