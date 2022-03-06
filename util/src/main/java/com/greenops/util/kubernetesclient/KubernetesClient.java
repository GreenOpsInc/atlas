package com.greenops.util.kubernetesclient;

import com.greenops.util.datamodel.git.GitCred;
import io.kubernetes.client.models.V1Secret;

public interface KubernetesClient {

    boolean storeGitCred(GitCred gitCred, String name);

    GitCred fetchGitCred(String name);

    V1Secret fetchSecretData(String namespace, String name);

}
