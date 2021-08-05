package com.greenops.pipelinereposerver.api.model.git;

import com.greenops.util.datamodel.git.GitCred;
import com.greenops.util.datamodel.git.GitCredMachineUser;
import com.greenops.util.datamodel.git.GitCredOpen;
import com.greenops.util.datamodel.git.GitCredToken;
import org.junit.jupiter.api.Test;

import static com.greenops.pipelinereposerver.api.util.Util.getGitCredAccessibleFromGitCred;
import static org.junit.jupiter.api.Assertions.assertEquals;

public class GitCredTests {
    @Test
    public void testsMachineUserToString() {
        GitCred machineUser = new GitCredMachineUser("root", "admin");
        var machineUserString = getGitCredAccessibleFromGitCred(machineUser).convertGitCredToString("https://gitcred.git");
        assertEquals("https://root:admin@gitcred.git", machineUserString);
    }

    @Test
    public void testGitCredTokenToString() {
        GitCred token = new GitCredToken("token");
        var tokenString = getGitCredAccessibleFromGitCred(token).convertGitCredToString("https://gitcred.git");
        assertEquals("https://token@gitcred.git", tokenString);
    }

    @Test
    public void testGitCredOpenToString() {
        GitCred open = new GitCredOpen();
        assertEquals("https://gitcred.git", getGitCredAccessibleFromGitCred(open).convertGitCredToString("https://gitcred.git"));
    }
}
