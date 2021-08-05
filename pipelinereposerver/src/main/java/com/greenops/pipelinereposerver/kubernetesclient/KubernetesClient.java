package com.greenops.pipelinereposerver.kubernetesclient;

import com.greenops.util.datamodel.git.GitCred;

public interface KubernetesClient {

    public boolean storeGitCred(GitCred gitCred, String name);

    public GitCred fetchGitCred(String name);
}


