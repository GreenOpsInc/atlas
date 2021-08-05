package com.greenops.pipelinereposerver.repomanager;


import com.greenops.pipelinereposerver.api.model.git.GitCredMachineUser;
import com.greenops.pipelinereposerver.api.model.git.GitCredOpen;
import com.greenops.pipelinereposerver.api.model.git.GitCredToken;
import com.greenops.pipelinereposerver.api.model.git.GitRepoSchema;
import com.greenops.pipelinereposerver.dbclient.DbClient;
import com.greenops.pipelinereposerver.kubernetesclient.KubernetesClient;
import org.apache.tomcat.util.http.fileupload.FileUtils;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.stream.Collectors;

import static org.junit.jupiter.api.Assertions.*;

public class RepoManagerTests {
    private RepoManager repoManager;
    private DbClient dbClient;
    private KubernetesClient kubernetesClient;
    private Path testFolderPath;
    private final String oldCommitHashPrivate = "43ffa3176fe4442b68eb52300721b03916f35fdd";
    private final String latestCommitHashPrivate = "36773475e7d293f94d2c26db5ec7b1cfb9ad0458";

    private final String oldCommitHashPublic = "50fc45eb4f22d09039e160c4c477d8cc1dc55912";
    private final String latestCommitHashPublic = "918405d162384c4a8680f98666cefea6821319c5";
    private final String token = "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";
    private final String username = "dummyaccount-test";
    private final String pwd = "FakeAcc123";

    private final String yamlContents = "# Build a service with Dockerfile\n" +
            "version: '1.0'\n" +
            "\n" +
            "steps:\n" +
            "  build-example:\n" +
            "    type: build\n" +
            "    description: yaml example\n" +
            "    image-name: codefresh-io/image\n" +
            "    dockerfile: Dockerfile\n" +
            "    tag: latest";


    @BeforeEach
    public void beforeEach() throws IOException {
        dbClient = Mockito.mock(DbClient.class);
        kubernetesClient = Mockito.mock(KubernetesClient.class);
        repoManager = new RepoManagerImpl(dbClient, kubernetesClient);
        testFolderPath = Paths.get(System.getProperty("user.dir"), repoManager.getOrgName(), "tmp");
    }

    @AfterEach
    public void afterEach() throws IOException {
        if (Files.exists(testFolderPath)) {
            FileUtils.deleteDirectory(testFolderPath.toFile());
        }
    }

    @Test
    public void testThatCloneWithCredOpenSucceeds() {
        GitRepoSchema schema = new GitRepoSchema("https://github.com/dummyaccount-test/testrepo2",
                "/",
                new GitCredOpen());
        assertTrue(repoManager.clone(schema));
        Path path = Paths.get(testFolderPath.toString(), "testrepo2");
        assertTrue(Files.exists(path));
        assertEquals(latestCommitHashPublic, repoManager.getLatestCommitFromCache("https://github.com/dummyaccount-test/testrepo2"));
        var repos = repoManager.getGitRepos().stream()
                .filter(gitRepoCache -> gitRepoCache.getGitRepoSchema().getGitRepo().equals("https://github.com/dummyaccount-test/testrepo2"))
                .collect(Collectors.toList());
        assertEquals(latestCommitHashPublic, repos.get(0).getRootCommitHash());
    }

    @Test
    public void testThatCloneWithCredTokenSucceeds() {
        GitRepoSchema schema = new GitRepoSchema("https://github.com/dummyaccount-test/testrepo",
                "/",
                new GitCredToken(token));
        assertTrue(repoManager.clone(schema));
        Path path = Paths.get(testFolderPath.toString(), "testrepo");
        assertTrue(Files.exists(path));
        assertEquals(latestCommitHashPrivate, repoManager.getLatestCommitFromCache("https://github.com/dummyaccount-test/testrepo"));
        var repos = repoManager.getGitRepos().stream()
                .filter(gitRepoCache -> gitRepoCache.getGitRepoSchema().getGitRepo().equals("https://github.com/dummyaccount-test/testrepo"))
                .collect(Collectors.toList());
        assertEquals(latestCommitHashPrivate, repos.get(0).getRootCommitHash());
    }

    @Test
    public void testThatCloneWithCredMachineUserSucceeds() {
        GitRepoSchema schema = new GitRepoSchema("https://github.com/dummyaccount-test/testrepo",
                "/",
                new GitCredMachineUser(username, pwd));
        assertTrue(repoManager.clone(schema));
        Path path = Paths.get(testFolderPath.toString(), "testrepo");
        assertTrue(Files.exists(path));
        assertEquals(latestCommitHashPrivate, repoManager.getLatestCommitFromCache("https://github.com/dummyaccount-test/testrepo"));
        var repos = repoManager.getGitRepos().stream()
                .filter(gitRepoCache -> gitRepoCache.getGitRepoSchema().getGitRepo().equals("https://github.com/dummyaccount-test/testrepo"))
                .collect(Collectors.toList());
        assertEquals(latestCommitHashPrivate, repos.get(0).getRootCommitHash());
    }


    @Test
    public void testDeleteWithCredOpenSucceeds() {
        GitRepoSchema schema = new GitRepoSchema("https://github.com/dummyaccount-test/testrepo2",
                "/",
                new GitCredOpen());
        assertTrue(repoManager.clone(schema));
        Path path = Paths.get(testFolderPath.toString(), "testrepo2");
        assertTrue(Files.exists(path));
        assertTrue(repoManager.delete(schema));
        assertFalse(Files.exists(path));
    }

    @Test
    public void testDeleteWithCredTokenSucceeds() {
        GitRepoSchema schema = new GitRepoSchema("https://github.com/dummyaccount-test/testrepo",
                "/",
                new GitCredToken(token));
        assertTrue(repoManager.clone(schema));
        Path path = Paths.get(testFolderPath.toString(), "testrepo");
        assertTrue(Files.exists(path));
        assertTrue(repoManager.delete(schema));
        assertFalse(Files.exists(path));
    }

    @Test
    public void testDeleteWithCredMachineUserSucceeds() {
        GitRepoSchema schema = new GitRepoSchema("https://github.com/dummyaccount-test/testrepo",
                "/",
                new GitCredMachineUser(username, pwd));
        assertTrue(repoManager.clone(schema));
        Path path = Paths.get(testFolderPath.toString(), "testrepo");
        assertTrue(Files.exists(path));
        assertTrue(repoManager.delete(schema));
        assertFalse(Files.exists(path));
    }


    @Test
    public void testSyncWithCredOpenSucceeds() {
        GitRepoSchema schema = new GitRepoSchema("https://github.com/dummyaccount-test/testrepo2",
                "/",
                new GitCredOpen());
        assertTrue(repoManager.clone(schema));
        Path pathFile1 = Paths.get(testFolderPath.toString(), "testrepo2", "testFile1");
        Path pathFile2 = Paths.get(testFolderPath.toString(), "testrepo2", "testFile2");
        assertTrue(Files.exists(pathFile1));
        assertTrue(Files.exists(pathFile2));
        repoManager.resetToVersion(oldCommitHashPublic, "https://github.com/dummyaccount-test/testrepo2");
        var repos = repoManager.getGitRepos().stream()
                .filter(gitRepoCache -> gitRepoCache.getGitRepoSchema().getGitRepo().equals("https://github.com/dummyaccount-test/testrepo2"))
                .collect(Collectors.toList());
        assertEquals(latestCommitHashPublic, repos.get(0).getRootCommitHash());
        assertTrue(Files.exists(pathFile1));
        assertFalse(Files.exists(pathFile2));
        assertTrue(repoManager.sync(schema));
        assertTrue(Files.exists(pathFile2));
        assertEquals(latestCommitHashPublic, repoManager.getLatestCommitFromCache("https://github.com/dummyaccount-test/testrepo2"));
        assertEquals(latestCommitHashPublic, repos.get(0).getRootCommitHash());

    }

    @Test
    public void testSyncWithCredTokenSucceeds() {
        GitRepoSchema schema = new GitRepoSchema("https://github.com/dummyaccount-test/testrepo",
                "/",
                new GitCredToken(token));
        assertTrue(repoManager.clone(schema));
        Path pathFile1 = Paths.get(testFolderPath.toString(), "testrepo", "testFile2");
        Path pathFile2 = Paths.get(testFolderPath.toString(), "testrepo", "codefresh-build-1.yml");
        assertTrue(Files.exists(pathFile1));
        assertTrue(Files.exists(pathFile2));
        repoManager.resetToVersion(oldCommitHashPrivate, "https://github.com/dummyaccount-test/testrepo");
        var repos = repoManager.getGitRepos().stream()
                .filter(gitRepoCache -> gitRepoCache.getGitRepoSchema().getGitRepo().equals("https://github.com/dummyaccount-test/testrepo"))
                .collect(Collectors.toList());
        assertEquals(latestCommitHashPrivate, repos.get(0).getRootCommitHash());
        assertTrue(Files.exists(pathFile1));
        assertFalse(Files.exists(pathFile2));
        assertTrue(repoManager.sync(schema));
        assertTrue(Files.exists(pathFile2));
        assertEquals(latestCommitHashPrivate, repoManager.getLatestCommitFromCache("https://github.com/dummyaccount-test/testrepo"));
        assertEquals(latestCommitHashPrivate, repos.get(0).getRootCommitHash());

    }

    @Test
    public void testSyncWithCredMachineUserSucceeds() {
        GitRepoSchema schema = new GitRepoSchema("https://github.com/dummyaccount-test/testrepo",
                "/",
                new GitCredMachineUser(username, pwd));
        assertTrue(repoManager.clone(schema));
        Path pathFile1 = Paths.get(testFolderPath.toString(), "testrepo", "testFile2");
        Path pathFile2 = Paths.get(testFolderPath.toString(), "testrepo", "codefresh-build-1.yml");
        assertTrue(Files.exists(pathFile1));
        assertTrue(Files.exists(pathFile2));
        repoManager.resetToVersion(oldCommitHashPrivate, "https://github.com/dummyaccount-test/testrepo");
        var repos = repoManager.getGitRepos().stream()
                .filter(gitRepoCache -> gitRepoCache.getGitRepoSchema().getGitRepo().equals("https://github.com/dummyaccount-test/testrepo"))
                .collect(Collectors.toList());
        assertEquals(latestCommitHashPrivate, repos.get(0).getRootCommitHash());
        assertTrue(Files.exists(pathFile1));
        assertFalse(Files.exists(pathFile2));
        assertTrue(repoManager.sync(schema));
        assertTrue(Files.exists(pathFile2));
        assertEquals(latestCommitHashPrivate, repoManager.getLatestCommitFromCache("https://github.com/dummyaccount-test/testrepo"));

    }

    @Test
    public void testContainsWorks() {
        GitRepoSchema schema = new GitRepoSchema("https://github.com/dummyaccount-test/testrepo",
                "/",
                new GitCredMachineUser(username, pwd));
        assertTrue(repoManager.clone(schema));
        assertTrue(repoManager.containsGitRepoSchema(schema));
    }

    @Test
    public void testResetToVersionWorks() {
        GitRepoSchema schema = new GitRepoSchema("https://github.com/dummyaccount-test/testrepo",
                "/",
                new GitCredMachineUser(username, pwd));
        assertTrue(repoManager.clone(schema));
        Path pathFile1 = Paths.get(testFolderPath.toString(), "testrepo", "testFile2");
        Path pathFile2 = Paths.get(testFolderPath.toString(), "testrepo", "codefresh-build-1.yml");
        assertTrue(Files.exists(pathFile1));
        assertTrue(Files.exists(pathFile2));
        repoManager.resetToVersion(oldCommitHashPrivate, "https://github.com/dummyaccount-test/testrepo");
        var repos = repoManager.getGitRepos().stream()
                .filter(gitRepoCache -> gitRepoCache.getGitRepoSchema().getGitRepo().equals("https://github.com/dummyaccount-test/testrepo"))
                .collect(Collectors.toList());
        assertEquals(latestCommitHashPrivate, repos.get(0).getRootCommitHash());
        assertTrue(Files.exists(pathFile1));
        assertFalse(Files.exists(pathFile2));
    }

    @Test
    public void testGetCurrentCommitWorks() {
        GitRepoSchema schema = new GitRepoSchema("https://github.com/dummyaccount-test/testrepo",
                "/",
                new GitCredMachineUser(username, pwd));
        assertTrue(repoManager.clone(schema));
        assertEquals(latestCommitHashPrivate, repoManager.getCurrentCommit(schema.getGitRepo()));
    }

    @Test
    public void testGetYamlFileContents() {
        GitRepoSchema schema = new GitRepoSchema("https://github.com/dummyaccount-test/testrepo",
                "/",
                new GitCredMachineUser(username, pwd));
        assertTrue(repoManager.clone(schema));
        assertEquals(yamlContents, repoManager.getYamlFileContents("https://github.com/dummyaccount-test/testrepo", "codefresh-build-1.yml"));

    }

}