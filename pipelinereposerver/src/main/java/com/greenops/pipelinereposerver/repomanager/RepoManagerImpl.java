package com.greenops.pipelinereposerver.repomanager;

import com.greenops.pipelinereposerver.dbclient.DbKey;
import com.greenops.pipelinereposerver.kubernetesclient.KubernetesClient;
import com.greenops.util.datamodel.git.GitRepoSchema;
import com.greenops.util.dbclient.DbClient;
import lombok.extern.slf4j.Slf4j;
import org.apache.logging.log4j.util.Strings;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
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

    private final String directory = "tmp";

    private final String orgName;
    private Set<GitRepoCache> gitRepos;

    @Autowired
    public RepoManagerImpl(DbClient dbClient, KubernetesClient kubernetesClient, @Value("${application.org-name}") String orgName) {
        this.gitRepos = new HashSet<>();
        this.orgName = orgName;
        if (!setupRepoCache(dbClient, kubernetesClient)) {
            throw new RuntimeException("The org's repos could not be cloned correctly. Please restart to try again.");
        }
        dbClient.shutdown();
    }

    @Override
    public String getOrgName() {
        return orgName;
    }

    @Override
    public Set<GitRepoCache> getGitRepos() {
        return gitRepos;
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
                log.info("Cloning repo {} was not successful. Error: {}\nCleaning up...", gitRepoSchema.getGitRepo(), new String(process.getErrorStream().readAllBytes()));
                delete(gitRepoSchema);
                return false;
            }
        } catch (IOException | InterruptedException e) {
            log.info("An error was thrown when attempting to clone the repo {}", gitRepoSchema.getGitRepo(), e);
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
        var newRepoCacheEntry = new GitRepoCache(oldGitRepoCache.getRootCommitHash(), getCurrentCommit(gitRepoSchema.getGitRepo()), gitRepoSchema);
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
            log.info("Too many git repos with the given url");
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
            var data = reader.lines().collect(Collectors.joining("\n"));
            reader.close();
            return data;
        } catch (IOException ex) {
            log.info("File was not found or could not be read.");
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
                var commitHash = getCurrentCommit(cachedGitRepoSchema.getGitRepo());
                if (commitHash == null) return false;
                gitRepos = gitRepos.stream().filter(gitRepoCache -> !gitRepoCache.getGitRepoSchema().getGitRepo().equals(gitRepoSchema.getGitRepo())).collect(Collectors.toSet());
                var updatedCacheSchema = listOfGitRepos.get(0);
                updatedCacheSchema.setCurrentCommitHash(commitHash);
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
                updatedCacheSchema.setCurrentCommitHash(gitCommit);
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
        return listOfGitRepos.get(0).getCurrentCommitHash();
    }

    @Override
    public boolean containsGitRepoSchema(GitRepoSchema gitRepoSchema) {
        return gitRepos.stream().anyMatch(gitRepoCache -> gitRepoCache.getGitRepoSchema().getGitRepo().equals(gitRepoSchema.getGitRepo()));
    }

    @Override
    public String getRootCommit(String gitRepoUrl) {
        var listOfGitRepos = gitRepos.stream().filter(gitRepoCache ->
                gitRepoCache.getGitRepoSchema().getGitRepo().equals(gitRepoUrl)).collect(Collectors.toList());
        if (listOfGitRepos.size() != 1) {
            //The size should never be greater than 1
            return null;
        }
        return listOfGitRepos.get(0).getRootCommitHash();
    }

    @Override
    public String getCurrentCommit(String gitRepoUrl) {
        try {
            var command = new CommandBuilder()
                    .gitLog(1, true)
                    .build();
            var process = new ProcessBuilder()
                    .command("/bin/bash", "-c", command)
                    .directory(new File(orgName + "/" + directory + "/" + getFolderName(gitRepoUrl)))
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

    private boolean setupRepoCache(DbClient dbClient, KubernetesClient kubernetesClient) {
        var command = new CommandBuilder()
                .mkdir(orgName)
                .cd(orgName)
                .mkdir(directory)
                .cd(directory)
                .build();
        try {
            var process = new ProcessBuilder()
                    .command("/bin/bash", "-c", command)
                    .start();
            int exitCode = process.waitFor();
            if (exitCode != 0) {
                log.info("Errors: {}", new String(process.getErrorStream().readAllBytes()));
                return false;
            }
        } catch (IOException | InterruptedException e) {
            return false;
        }
        var listOfTeams = dbClient.fetchStringList(DbKey.makeDbListOfTeamsKey(orgName));
        if (listOfTeams == null) {
            log.info("No teams in org {}", orgName);
            return true;
        }
        log.info("Fetched all teams and cloning pipeline repos for org {}", orgName);
        for (var teamName : listOfTeams) {
            var teamSchema = dbClient.fetchTeamSchema(DbKey.makeDbTeamKey(orgName, teamName));
            if (teamSchema == null) {
                log.info("The team {} doesn't exist, so cloning will be skipped", teamName);
                continue;
            }
            for (var pipelineSchema : teamSchema.getPipelineSchemas()) {
                var gitRepoSchema = pipelineSchema.getGitRepoSchema();
                var secretName = DbKey.makeSecretName(teamSchema.getOrgName(), teamName, pipelineSchema.getPipelineName());
                var gitCred = kubernetesClient.fetchGitCred(secretName);
                if (gitCred == null) {
                    throw new RuntimeException("Could not recover gitCred from kubernetes secrets");
                }
                gitRepoSchema.setGitCred(gitCred);
                if (!clone(gitRepoSchema)) {
                    return false;
                }
            }
            log.info("Finished cloning pipeline repos for team {}", teamName);
        }
        return true;
    }
}
