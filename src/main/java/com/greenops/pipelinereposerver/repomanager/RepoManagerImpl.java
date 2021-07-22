package com.greenops.pipelinereposerver.repomanager;

import com.greenops.pipelinereposerver.api.model.git.GitRepoSchema;
import com.greenops.pipelinereposerver.dbclient.DbClient;
import com.greenops.pipelinereposerver.dbclient.DbKey;
import lombok.extern.slf4j.Slf4j;
import org.apache.logging.log4j.util.Strings;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.HashSet;
import java.util.List;
import java.util.Set;
import java.util.stream.Collectors;

import static com.greenops.pipelinereposerver.repomanager.CommandBuilder.getFolderName;

@Slf4j
@Component
public class RepoManagerImpl implements RepoManager {

    private final String orgName = "org"; //TODO: Needs to be updated when we decide how to configure organization
    private final String directory = "tmp";

    private Set<GitRepoCache> gitRepos;

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
                    .cd(getFolderName(gitRepoSchema.getGitRepo()))
                    .gitLog(1, true)
                    .build();
            var process = new ProcessBuilder()
                    .command("/bin/bash", "-c", command)
                    .directory(new File(orgName + "/" + directory))
                    .start();
            int exitCode = process.waitFor();
            if (exitCode == 0) {
                log.info("Cloning repo {} was successful.", gitRepoSchema.getGitRepo());
                var commitHash = new String(process.getInputStream().readAllBytes()).split("\n")[0];
                gitRepos.add(new GitRepoCache(commitHash, commitHash, gitRepoSchema));
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
        var oldGitRepoCache = gitRepos.stream().filter(gitRepoCache -> gitRepoCache.getGitRepoSchema().getGitRepo().equals(gitRepoSchema.getGitRepo())).findFirst().orElse(null);
        if (oldGitRepoCache == null) {
            return false;
        }

        gitRepos = gitRepos.stream().filter(gitRepoCache -> !gitRepoCache.getGitRepoSchema().getGitRepo().equals(gitRepoSchema.getGitRepo())).collect(Collectors.toSet());
        var newRepoCacheEntry = new GitRepoCache(oldGitRepoCache.getRootCommitHash(), getCurrentCommit(gitRepoSchema), gitRepoSchema);
        gitRepos.add(newRepoCacheEntry);
        if (sync(gitRepoSchema)) {
            return true;
        } else {
            gitRepos.remove(newRepoCacheEntry);
            gitRepos.add(oldGitRepoCache);
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
                gitRepos = gitRepos.stream().filter(gitRepoCache -> !gitRepoCache.getGitRepoSchema().getGitRepo().equals(gitRepoSchema.getGitRepo())).collect(Collectors.toSet());
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
                .filter(gitRepoCache -> gitRepoCache.getGitRepoSchema().getGitRepo().equals(gitRepoUrl))
                .collect(Collectors.toList());
        if (listOfSchemas.size() != 1) {
            log.error("Too many git repos with the given url");
            return null;
        }

        try {
            var pathToRoot = strip(listOfSchemas.get(0).getGitRepoSchema().getPathToRoot(), '/');
            var truncatedFilePath = strip(filename, '/');
            var path = Paths.get(Strings.join(
                    List.of(orgName, directory, getFolderName(gitRepoUrl), pathToRoot, truncatedFilePath),
                    '/')
            );
            var reader = Files.newBufferedReader(path);
            return reader.lines().collect(Collectors.joining("\n"));
        } catch (IOException ex) {
            log.error("File was not found or could not be read.");
            return null;
        }
    }

    @Override
    public boolean sync(GitRepoSchema gitRepoSchema) {
        var listOfGitRepos = gitRepos.stream().filter(gitRepoCache ->
                gitRepoCache.getGitRepoSchema().getGitRepo().equals(gitRepoSchema.getGitRepo())).collect(Collectors.toList());
        if (listOfGitRepos.size() != 1) {
            //The size should never be greater than 1
            return false;
        }
        var cachedGitRepoSchema = listOfGitRepos.get(0).getGitRepoSchema();
        try {
            var command = new CommandBuilder()
                    .gitCheckout("main")
                    .gitPull(cachedGitRepoSchema)
                    .build();
            var process = new ProcessBuilder()
                    .command("/bin/bash", "-c", command)
                    .directory(new File(orgName + "/" + directory + "/" + getFolderName(cachedGitRepoSchema.getGitRepo())))
                    .start();
            int exitCode = process.waitFor();
            if (exitCode == 0) {
                log.info("Pulling repo {} was successful.", cachedGitRepoSchema.getGitRepo());
                var commitHash = getCurrentCommit(cachedGitRepoSchema);
                if (commitHash == null) return false;
                gitRepos = gitRepos.stream().filter(gitRepoCache -> !gitRepoCache.getGitRepoSchema().getGitRepo().equals(gitRepoSchema.getGitRepo())).collect(Collectors.toSet());
                var updatedCacheSchema = listOfGitRepos.get(0);
                updatedCacheSchema.addCommitHashToHistory(commitHash);
                updatedCacheSchema.setRootCommitHash(commitHash);
                gitRepos.add(updatedCacheSchema);
                return true;
            } else {
                log.info("Pulling repo {} was not successful.", cachedGitRepoSchema.getGitRepo());
                return false;
            }
        } catch (IOException | InterruptedException e) {
            log.error("An error was thrown when attempting to sync the repo {}", cachedGitRepoSchema.getGitRepo(), e);
            delete(cachedGitRepoSchema);
            return false;
        }
    }

    @Override
    public boolean resetToVersion(String gitCommit, String gitRepoUrl) {
        var listOfGitRepos = gitRepos.stream().filter(gitRepoCache ->
                gitRepoCache.getGitRepoSchema().getGitRepo().equals(gitRepoUrl)).collect(Collectors.toList());
        if (listOfGitRepos.size() != 1) {
            //The size should never be greater than 1
            return false;
        }
        var cachedGitRepoSchema = listOfGitRepos.get(0).getGitRepoSchema();
        try {
            var command = new CommandBuilder()
                    .gitCheckout(gitCommit)
                    .build();
            var process = new ProcessBuilder()
                    .command("/bin/bash", "-c", command)
                    .directory(new File(orgName + "/" + directory + "/" + getFolderName(cachedGitRepoSchema.getGitRepo())))
                    .start();
            int exitCode = process.waitFor();
            if (exitCode == 0) {
                gitRepos = gitRepos.stream().filter(gitRepoCache -> !gitRepoCache.getGitRepoSchema().getGitRepo().equals(gitRepoUrl)).collect(Collectors.toSet());
                var updatedCacheSchema = listOfGitRepos.get(0);
                updatedCacheSchema.addCommitHashToHistory(gitCommit);
                gitRepos.add(updatedCacheSchema);
                log.info("Updating repo version {} was successful.", cachedGitRepoSchema.getGitRepo());
                return true;
            } else {
                log.info("Updating repo version {} was not successful.", cachedGitRepoSchema.getGitRepo());
                return false;
            }
        } catch (IOException | InterruptedException e) {
            log.error("An error was thrown when attempting to update the repo version {}", cachedGitRepoSchema.getGitRepo(), e);
            return false;
        }
    }

    @Override
    public String getLatestCommitFromCache(String gitRepoUrl) {
        var listOfGitRepos = gitRepos.stream().filter(gitRepoCache ->
                gitRepoCache.getGitRepoSchema().getGitRepo().equals(gitRepoUrl)).collect(Collectors.toList());
        if (listOfGitRepos.size() != 1) {
            //The size should never be greater than 1
            return null;
        }
        return listOfGitRepos.get(0).getCommitHashHistory().get(0);
    }

    @Override
    public boolean containsGitRepoSchema(GitRepoSchema gitRepoSchema) {
        return gitRepos.stream().anyMatch(gitRepoCache -> gitRepoCache.getGitRepoSchema().getGitRepo().equals(gitRepoSchema.getGitRepo()));
    }

    private String getCurrentCommit(GitRepoSchema gitRepoSchema) {
        try {
            var command = new CommandBuilder()
                    .gitLog(1, true)
                    .build();
            var process = new ProcessBuilder()
                    .command("/bin/bash", "-c", command)
                    .directory(new File(orgName + "/" + directory + "/" + getFolderName(gitRepoSchema.getGitRepo())))
                    .start();
            int exitCode = process.waitFor();
            if (exitCode == 0) {
                log.info("Fetching commit was successful.");
                return new String(process.getInputStream().readAllBytes()).split("\n")[0];
            } else {
                log.info("Fetching commit was not successful.");
                return null;
            }
        } catch (IOException | InterruptedException e) {
            log.error("An error was thrown when attempting to fetch the commit hash", e);
            return null;
        }
    }

    private String strip(String str, char delimiter) {
        var beg = 0;
        var end = str.length() - 1;
        while (beg < str.length() && str.charAt(beg) == delimiter) {
            beg++;
        }
        while (end >= 0 && str.charAt(end) == delimiter) {
            end--;
        }
        if (end < beg) {
            return "";
        } else {
            return str.substring(beg, end + 1);
        }
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
            }
            log.info("Finished cloning pipeline repos for team {}", teamName);
        }
        return true;
    }
}
