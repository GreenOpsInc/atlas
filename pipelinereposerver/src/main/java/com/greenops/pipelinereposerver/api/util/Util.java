package com.greenops.pipelinereposerver.api.util;

import com.greenops.util.datamodel.git.GitCred;
import com.greenops.util.datamodel.git.GitCredAccessible;

public class Util {
    public static GitCredAccessible getGitCredAccessibleFromGitCred(GitCred gitCred) {
        return (GitCredAccessible) gitCred;
    }
}
