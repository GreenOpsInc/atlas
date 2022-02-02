package com.greenops.util.tslmanager;

import com.greenops.util.kubernetesclient.KubernetesClient;
import io.kubernetes.client.models.V1Secret;
import lombok.extern.slf4j.Slf4j;
import org.apache.kafka.clients.CommonClientConfigs;
import org.apache.kafka.common.config.SslConfigs;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.io.File;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.util.HashMap;
import java.util.Map;

@Slf4j
@Component
public class TLSManagerImpl implements TLSManager {
    
    private static final String NAMESPACE = "default";
    private static final String KAFKA_SECRET_NAME = "kafka-tls";
    private static final String SECRET_KAFKA_KEYSTORE_NAME = "kafka.keystore.jks";
    private static final String SECRET_KAFKA_TRUSTSTORE_NAME = "kafka.truststore.jks";
    private static final String SECRET_KAFKA_KEY_CREDENTIALS = "kafka.key.credentials";
    private static final String SECRET_KAFKA_KEYSTORE_CREDENTIALS = "kafka.keystore.credentials";
    private static final String SECRET_KAFKA_TRUSTSTORE_CREDENTIALS = "kafka.truststore.credentials";
    private final KubernetesClient kclient;

    @Autowired
    public TLSManagerImpl(KubernetesClient kclient) {
        this.kclient = kclient;
    }

    @Override
    public Map<String, Object> getKafkaSSLConfProps(String keystoreLocation, String truststoreLocation) throws Exception {
        File keyStoreFile = new File(keystoreLocation);
        File truststoreFile = new File(truststoreLocation);
        File parent = keyStoreFile.getParentFile();
        if (parent != null && !parent.exists() && !parent.mkdirs()) {
            throw new IllegalStateException("Couldn't create .atlas config directory: " + parent);
        }

        if (keyStoreFile.exists()) {
            keyStoreFile.delete();
        }
        if (truststoreFile.exists()) {
            truststoreFile.delete();
        }

        V1Secret secret = this.kclient.fetchSecretData(NAMESPACE, KAFKA_SECRET_NAME);
        if (secret == null) {
            return null;
        }

        byte[] keystoreData = secret.getData().get(SECRET_KAFKA_KEYSTORE_NAME);
        byte[] truststoreData = secret.getData().get(SECRET_KAFKA_TRUSTSTORE_NAME);
        String kafkaKeystoreCreds = new String(secret.getData().get(SECRET_KAFKA_KEYSTORE_CREDENTIALS), StandardCharsets.UTF_8);
        String kafkaTruststoreCreds = new String(secret.getData().get(SECRET_KAFKA_TRUSTSTORE_CREDENTIALS), StandardCharsets.UTF_8);
        String kafkaKeyCreds = new String(secret.getData().get(SECRET_KAFKA_KEY_CREDENTIALS), StandardCharsets.UTF_8);

        keyStoreFile = new File(keystoreLocation);
        keyStoreFile.createNewFile();
        Files.write(keyStoreFile.toPath(), keystoreData);

        truststoreFile = new File(truststoreLocation);
        truststoreFile.createNewFile();
        Files.write(truststoreFile.toPath(), truststoreData);

        Map<String, Object> configProps = new HashMap<>();
        configProps.put(CommonClientConfigs.SECURITY_PROTOCOL_CONFIG, "SSL");
        configProps.put(SslConfigs.SSL_KEYSTORE_LOCATION_CONFIG, keystoreLocation);
        configProps.put(SslConfigs.SSL_TRUSTSTORE_LOCATION_CONFIG, truststoreLocation);
        configProps.put(SslConfigs.SSL_KEYSTORE_PASSWORD_CONFIG, kafkaKeystoreCreds);
        configProps.put(SslConfigs.SSL_TRUSTSTORE_PASSWORD_CONFIG, kafkaTruststoreCreds);
        configProps.put(SslConfigs.SSL_KEY_PASSWORD_CONFIG, kafkaKeyCreds);
        return configProps;
    }
}
