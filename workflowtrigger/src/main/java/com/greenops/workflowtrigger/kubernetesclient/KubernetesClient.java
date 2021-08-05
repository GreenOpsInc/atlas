package com.greenops.workflowtrigger.kubernetesclient;

import com.greenops.workflowtrigger.api.model.git.GitCred;

public interface KubernetesClient {

    public boolean storeGitCred(GitCred gitCred, String name);

    public GitCred fetchGitCred(String name);
}


