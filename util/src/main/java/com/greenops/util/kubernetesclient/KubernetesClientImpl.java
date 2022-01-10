package com.greenops.util.kubernetesclient;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.util.datamodel.git.GitCred;
import io.kubernetes.client.ApiException;
import io.kubernetes.client.Configuration;
import io.kubernetes.client.apis.CoreV1Api;
import io.kubernetes.client.informer.ResourceEventHandler;
import io.kubernetes.client.informer.SharedIndexInformer;
import io.kubernetes.client.informer.SharedInformerFactory;
import io.kubernetes.client.models.V1Namespace;
import io.kubernetes.client.models.V1ObjectMeta;
import io.kubernetes.client.models.V1Secret;
import io.kubernetes.client.models.V1SecretList;
import io.kubernetes.client.util.CallGeneratorParams;
import io.kubernetes.client.util.ClientBuilder;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.io.IOException;

@Slf4j
@Component
public class KubernetesClientImpl implements KubernetesClient {
    private static final String SECRETS_KEY_NAME = "data";
    private static final String GITCRED_NAMESPACE = "gitcred";

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
    public boolean storeGitCred(GitCred gitCred, String name) {
        return storeSecret(gitCred, GITCRED_NAMESPACE, name);
    }

    @Override
    public GitCred fetchGitCred(String name) {
        var secret = readSecret(GITCRED_NAMESPACE, name);
        var gitCredData = secret.getData().get(SECRETS_KEY_NAME);
        try {
            if (gitCredData == null) {
                return null;
            }
            return objectMapper.readValue(gitCredData, GitCred.class);
        } catch (IOException e) {
            throw new RuntimeException("Could not deserialize gitCred.", e);
        }
    }

    @Override
    public V1Secret fetchSecretData(String namespace, String name) {
        return readSecret(namespace, name);
    }

    @Override
    public void watchSecretData(String name, String namespace, WatchSecretHandler handler) {
        SharedInformerFactory factory = new SharedInformerFactory();
        CoreV1Api coreV1Api = new CoreV1Api();

        SharedIndexInformer<V1Secret> informer = factory.sharedIndexInformerFor(
                (CallGeneratorParams params) -> {
                    try {
                        return coreV1Api.readNamespacedSecretCall(name, namespace, null, null, null, null, null);
                    } catch (ApiException e) {
                        e.printStackTrace();
                        throw new RuntimeException("Could not initialize Kubernetes Client Secret watcher", e);
                    }
                },
                V1Secret.class,
                V1SecretList.class);

        informer.addEventHandler(
                new ResourceEventHandler<>() {
                    @Override
                    public void onAdd(V1Secret secret) {
                        handler.handle(secret);
                    }

                    @Override
                    public void onUpdate(V1Secret oldSecret, V1Secret newSecret) {
                        handler.handle(newSecret);
                    }

                    @Override
                    public void onDelete(V1Secret secret, boolean deletedFinalStateUnknown) {
                        handler.handle(null);
                    }
                });

        factory.startAllRegisteredInformers();
    }

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

    private V1Secret readSecret(String namespace, String name) {
        try {
            var secret = api.readNamespacedSecret(name, namespace, null, null, null);
            return secret;
        } catch (ApiException e) {
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
