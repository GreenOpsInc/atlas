package com.greenops.util.kubernetesclient;

import io.kubernetes.client.models.V1Secret;

public interface WatchSecretHandler {
    void handle(V1Secret secret);
}
