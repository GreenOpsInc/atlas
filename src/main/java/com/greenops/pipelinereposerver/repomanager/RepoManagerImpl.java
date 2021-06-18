package com.greenops.pipelinereposerver.repomanager;

import com.greenops.pipelinereposerver.api.model.git.GitRepoSchema;
import com.greenops.pipelinereposerver.dbclient.DbClient;
import com.greenops.pipelinereposerver.dbclient.DbKey;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.HashSet;
import java.util.Optional;
import java.util.Set;
import java.util.stream.Collectors;

import static com.greenops.pipelinereposerver.repomanager.CommandBuilder.getFolderName;

@Slf4j
@Component
public class RepoManagerImpl implements RepoManager {

    private final String orgName = "temporary"; //TODO: Needs to be updated when we decide how to configure organization
    private final String directory = "tmp";

    private Set<GitRepoSchema> gitRepos;

    @Autowired
    public RepoManagerImpl(DbClient dbClient) {
        this.gitRepos = new HashSet<>();
        if (!setupGitCli()) {
            throw new RuntimeException("Git could not be installed. Something may be wrong with the configuration.");
        }
        if (!setupRepoCache(dbClient)) {
            throw new RuntimeException("The org's repos could not be cloned correctly. Please restart to try again.");
        }
        dbClient.shutdown();
    }

    @Override
    public String getOrgName() {
        return orgName;
    }

    @Override
    public boolean clone(GitRepoSchema gitRepoSchema) {
        if (containsGitRepoSchema(gitRepoSchema)) {
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
                gitRepos.add(gitRepoSchema);
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
    public boolean update(GitRepoSchema gitRepoSchema) {
        //There should only be one answer that fits the filter for oldGitRepoSchema
        var oldGitRepoSchema = gitRepos.stream().filter(gitRepoSchema1 -> gitRepoSchema1.getGitRepo().equals(gitRepoSchema.getGitRepo())).findFirst().orElse(null);

        gitRepos = gitRepos.stream().filter(gitRepoSchema1 -> !gitRepoSchema1.getGitRepo().equals(gitRepoSchema.getGitRepo())).collect(Collectors.toSet());
        gitRepos.add(gitRepoSchema);
        if (sync(gitRepoSchema)) {
            return true;
        } else {
            gitRepos.remove(gitRepoSchema);
            gitRepos.add(oldGitRepoSchema);
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
            if (exitCode == 0) {
                gitRepos.remove(gitRepoSchema);
                return true;
            } else {
                return false;
            }
        } catch (IOException | InterruptedException e) {
            log.error("An error was thrown when attempting to delete the repo {}", gitRepoSchema.getGitRepo(), e);
            return false;
        }
    }

    @Override
    public String getYamlFileContents(String gitRepoUrl, String filename) {
        var listOfSchemas = gitRepos.stream()
                .filter(gitRepoSchema -> gitRepoSchema.getGitRepo().equals(gitRepoUrl))
                .collect(Collectors.toList());
        if (listOfSchemas.size() != 1) {
            throw new RuntimeException("Too many git repos with the given url");
        }

        try {
            Optional<Path> fileToFind = Files.walk(Paths.get(listOfSchemas.get(0).getPathToRoot()))
                    .filter(file -> Files.isRegularFile(file) && file.getFileName().equals(filename) && (file.endsWith("yaml") || file.endsWith("yml")))
                    .findFirst();
            if (fileToFind.isPresent()) {
                log.info("Found the file : {}", fileToFind.get().getFileName());
                return Files.readString(fileToFind.get());
            }
        } catch (IOException ex) {
            throw new RuntimeException("File was not found.");
        }

        return null;
    }

    @Override
    public boolean sync(GitRepoSchema gitRepoSchema) {
        var listOfGitRepos = gitRepos.stream().filter(gitRepoSchema1 ->
                gitRepoSchema1.getGitRepo().equals(gitRepoSchema.getGitRepo())).collect(Collectors.toList());
        if (listOfGitRepos.size() != 1) {
            //The size should never be greater than 1
            return false;
        }
        var cachedGitRepoSchema = listOfGitRepos.get(0);
        try {
            var command = new CommandBuilder()
                    .gitPull(cachedGitRepoSchema)
                    .build();
            var process = new ProcessBuilder()
                    .command("/bin/bash", "-c", command)
                    .directory(new File(orgName + "/" + directory + "/" + getFolderName(cachedGitRepoSchema.getGitRepo())))
                    .start();
            int exitCode = process.waitFor();
            if (exitCode == 0) {
                log.info("Pulling repo {} was successful.", cachedGitRepoSchema.getGitRepo());
                return true;
            } else {
                log.info("Pulling repo {} was not successful.", cachedGitRepoSchema.getGitRepo());
                return false;
            }
        } catch (IOException | InterruptedException e) {
            log.error("An error was thrown when attempting to clone the repo {}", cachedGitRepoSchema.getGitRepo(), e);
            delete(cachedGitRepoSchema);
            return false;
        }
    }

    @Override
    public boolean containsGitRepoSchema(GitRepoSchema gitRepoSchema) {
        return gitRepos.stream().anyMatch(gitRepoSchema1 -> gitRepoSchema1.getGitRepo().equals(gitRepoSchema.getGitRepo()));
    }

    private boolean setupGitCli() {
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

    private boolean setupRepoCache(DbClient dbClient) {
        var listOfTeams = dbClient.fetchList(DbKey.makeDbListOfTeamsKey(orgName));
        if (listOfTeams == null) {
            log.info("No teams in org {}", orgName);
            return true;
        }
        log.info("Fetched all teams and cloning pipeline repos for org {}", orgName);
        for (var teamName : listOfTeams) {
            var teamSchema = dbClient.fetchTeamSchema(DbKey.makeDbTeamKey(orgName, teamName));
            if (teamSchema == null) {
                log.error("The team {} doesn't exist, so cloning will be skipped", teamName);
                continue;
            }
            for (var pipelineSchema : teamSchema.getPipelineSchemas()) {
                if (!clone(pipelineSchema.getGitRepoSchema())) {
                    return false;
                }
                gitRepos.add(pipelineSchema.getGitRepoSchema());
            }
            log.info("Finished cloning pipeline repos for team {}", teamName);
        }
        return true;
    }
}
