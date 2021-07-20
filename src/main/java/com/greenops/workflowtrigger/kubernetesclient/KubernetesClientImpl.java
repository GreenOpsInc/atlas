package com.greenops.workflowtrigger.kubernetesclient;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;

import io.kubernetes.client.ApiException;
import io.kubernetes.client.Configuration;
import io.kubernetes.client.apis.CoreV1Api;
import io.kubernetes.client.models.V1Namespace;
import io.kubernetes.client.models.V1ObjectMeta;
import io.kubernetes.client.models.V1Secret;
import io.kubernetes.client.util.ClientBuilder;
import java.io.IOException;

import com.greenops.workflowtrigger.api.model.git.GitCred;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

@Slf4j
@Component
public class KubernetesClientImpl implements KubernetesClient {
    private static final String SECRETS_KEY_NAME = "data";

    private final CoreV1Api api;
    private final ObjectMapper objectMapper;

    @Autowired
    public KubernetesClientImpl(ObjectMapper objectMapper) throws IOException {
        var client = ClientBuilder.cluster().build();
        Configuration.setDefaultApiClient(client);
        api = new CoreV1Api();
        this.objectMapper = objectMapper;
    }

    @Override
    public boolean storeSecret(Object object, String namespace, String name) {
        try {
            api.readNamespace(namespace, null, null, null);
        } catch (ApiException readNamespaceException) {
            var v1Namespace = new V1Namespace();
            var v1MetaData = new V1ObjectMeta();
            v1MetaData.setName(namespace);
            v1Namespace.setMetadata(v1MetaData);
            try {
                api.createNamespace(v1Namespace, null, null, null);
            } catch (ApiException createNamespaceException) {
                log.error(String.format("Failed to create namespace. Error: %s", createNamespaceException.getResponseBody()), createNamespaceException);
                return false;
            }
        }
        if (object == null) {
            return deleteSecret(namespace, name);
        }
        if (readSecret(namespace, name) == null) {
            return createSecret(object, namespace, name);
        }
        return updateSecret(object, namespace, name);
    }

    @Override
    public Object readSecret(String namespace, String name) {
        try {
            var secret = api.readNamespacedSecret(name, namespace, null, null, null);
            var gitCredData = secret.getData().get(SECRETS_KEY_NAME);
            try {
                return objectMapper.readValue(gitCredData, GitCred.class);
            } catch (Exception e) {
                throw new RuntimeException("Could not deserialize gitCred.", e);
            }
        } catch (ApiException e){
            log.error(String.format("Failed to read secret. Error: %s", e.getResponseBody()), e);
        }
        return null;
    }

    private V1Secret makeSecret(Object object, String namespace, String name) {
        var secret = new V1Secret();
        secret.setApiVersion("v1");
        secret.setKind("Secret");
        secret.setType("Opaque");
        var metadata = new V1ObjectMeta();
        metadata.setName(name);
        metadata.setNamespace(namespace);
        secret.setMetadata(metadata);
        try {
            var gitCredString = objectMapper.writeValueAsString(object);
            secret.putStringDataItem(SECRETS_KEY_NAME, gitCredString);
        } catch (JsonProcessingException e) {
            throw new RuntimeException("Serializing gitCred failed.", e);
        }
        return secret;
    }

    private boolean createSecret(Object object, String namespace, String name) {
        var secret = makeSecret(object, namespace, name);
        try {
            api.createNamespacedSecret(namespace, secret, null, null, null);
            return true;
        } catch (ApiException e) {
            log.error(String.format("Failed to create secret. Error: %s", e.getResponseBody()), e);
            return false;
        }
    }

    private boolean updateSecret(Object object, String namespace, String name) {
        var secret = makeSecret(object, namespace, name);
        try {
            api.replaceNamespacedSecret(name, namespace, secret, null, null);
            return true;
        } catch (ApiException e) {
            log.error(String.format("Failed to update secret. Error: %s", e.getResponseBody()), e);
            return false;
        }
    }

    private boolean deleteSecret(String namespace, String name) {
        try {
            api.deleteNamespacedSecret(name, namespace, null, null, null, null, null, null);
            return true;
        } catch (ApiException e) {
            log.error(String.format("Failed to delete secret. Error: %s", e.getResponseBody()), e);
            return false;
        }
    }
}
