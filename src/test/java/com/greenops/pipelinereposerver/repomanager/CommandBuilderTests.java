package com.greenops.pipelinereposerver.repomanager;

import com.greenops.pipelinereposerver.api.model.git.GitCredMachineUser;
import com.greenops.pipelinereposerver.api.model.git.GitRepoSchema;
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;


public class CommandBuilderTests {

    @Test
    void testMkdirCommand() {
        var builder = new CommandBuilder();
        assertEquals("mkdir testDirectory", builder.mkdir("testDirectory").build());
    }

    @Test
    void testStringMultipleCommands() {
        var builder = new CommandBuilder()
                .cd("testDirectory")
                .mkdir("testDirectory");
        assertEquals("cd testDirectory; mkdir testDirectory", builder.build());
    }

    @Test
    void testGitCloneCommand() {
        var testSchema = new GitRepoSchema("https://testRepo.git", "/root",
                new GitCredMachineUser("root", "admin"));
        var builder = new CommandBuilder();
        assertEquals("git clone https://root:admin@testRepo.git", builder.gitClone(testSchema).build());
    }

    @Test
    void testGitPullCommand() {
        var testSchema = new GitRepoSchema("https://testRepo.git", "/root",
                new GitCredMachineUser("root", "admin"));
        var builder = new CommandBuilder();
        assertEquals("git pull https://root:admin@testRepo.git", builder.gitPull(testSchema).build());
    }

    @Test
    void testGetFolderName() {
        var testSchema = new GitRepoSchema("https://testRepo.git", "/root",
                new GitCredMachineUser("root", "admin"));
        assertEquals("testRepo", CommandBuilder.getFolderName(testSchema.getGitRepo()));
    }

    @Test
    void testGetFolderNameWhenEndingWithBackSlash() {
        var testSchema = new GitRepoSchema("https://testRepo.git/", "/root",
                new GitCredMachineUser("root", "admin"));
        assertEquals("testRepo", CommandBuilder.getFolderName(testSchema.getGitRepo()));
    }

    @Test
    void testGetFolderNameWhenNoGitExtension() {
        var testSchema = new GitRepoSchema("https://testRepo/", "/root",
                new GitCredMachineUser("root", "admin"));
        assertEquals("testRepo", CommandBuilder.getFolderName(testSchema.getGitRepo()));
    }

    @Test
    void deleteFolderCommand() {
        var testSchema = new GitRepoSchema("https://testRepo.git", "/root",
                new GitCredMachineUser("root", "admin"));
        var builder = new CommandBuilder();
        assertEquals("rm -rf testRepo", builder.deleteFolder(testSchema).build());
    }
}
