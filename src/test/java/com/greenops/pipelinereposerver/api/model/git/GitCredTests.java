package com.greenops.pipelinereposerver.api.model.git;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertEquals;

public class GitCredTests {
    @Test
    public void testsMachineUserToString() {
        GitCred machineUser = new GitCredMachineUser("root", "admin");
        var machineUserString = machineUser.convertGitCredToString("https://gitcred.git");
        assertEquals("https://root:admin@gitcred.git", machineUserString);
    }

    @Test
    public void testGitCredTokenToString() {
        GitCred token = new GitCredToken("token");
        var tokenString = token.convertGitCredToString("https://gitcred.git");
        assertEquals("https://token@gitcred.git", tokenString);
    }

    @Test
    public void testGitCredOpenToString() {
        GitCred open = new GitCredOpen();
        assertEquals("https://gitcred.git", open.convertGitCredToString("https://gitcred.git"));
    }
}
