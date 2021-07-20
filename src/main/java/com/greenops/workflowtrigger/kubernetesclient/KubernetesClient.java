package com.greenops.workflowtrigger.kubernetesclient;

import com.greenops.workflowtrigger.api.model.git.GitCred;

public interface KubernetesClient {

    public boolean storeSecret(Object object, String namespace, String name);
    public Object readSecret(String namespace, String name);
}


